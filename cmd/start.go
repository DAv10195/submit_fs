package cmd

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nikitakogan1/submit-file-server/path"
	fileserver "github.com/nikitakogan1/submit-file-server/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func newStartCommand(ctx context.Context, args []string) *cobra.Command {
	var setupErr error
	startCmd := &cobra.Command{
		Use: start,
		Short: fmt.Sprintf("%s %s", start, submitFileServer),
		SilenceUsage: true,
		SilenceErrors: true,
		RunE: func (cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if setupErr != nil {
				return setupErr
			}
			logLevel := viper.GetString(flagLogLevel)
			level, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logrus.SetLevel(level)
			logFile := viper.GetString(flagLogFile)
			if logFile != "" {
				lumberjackLogger := &lumberjack.Logger{
					Filename:   viper.GetString(flagLogFile),
					MaxSize:    viper.GetInt(flagLogFileMaxSize),
					MaxBackups: viper.GetInt(flagLogFileMaxBackups),
					MaxAge:     viper.GetInt(flagLogFileMaxAge),
					LocalTime:  true,
				}
				if viper.GetBool(flagLogFileAndStdout) {
					logrus.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
				} else {
					logrus.SetOutput(lumberjackLogger)
				}
			} else {
				logger.Debug("log file undefined")
			}

			cfg := &fileserver.Config{}
			cfg.Port = viper.GetInt(flagFileServerPort)
			baseRouter := mux.NewRouter()
			baseRouter.Use(fileserver.AuthenticationMiddleware)
			router := fileserver.InitRouters(baseRouter)
			fileserver.InitFolders()
			fs := fileserver.NewFileServer(router)
			go func() {
				if err := fs.ListenAndServe(); err != http.ErrServerClosed {
					logger.WithError(err).Fatal("submit server crashed")
				}
			}()
			logger.Info("server is running")
			<- ctx.Done()
			logger.Info("stopping server...")
			ctx, timeout := context.WithTimeout(context.Background(), time.Minute)
			defer timeout()
			if err := fs.Shutdown(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	configFlagSet := pflag.NewFlagSet(fileServer, pflag.ContinueOnError)
	_ = configFlagSet.StringP(flagConfigFile, "c", "", "path to submit server config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(args[1:])
	configFilePath, _ := configFlagSet.GetString(flagConfigFile)
	if configFilePath == "" {
		configFilePath = filepath.Join(path.GetDefaultConfigFilePath(), defaultConfigFileName)
	}
	viper.SetConfigType(yaml)
	viper.SetConfigFile(configFilePath)
	viper.SetDefault(flagLogFileAndStdout, deLogFileAndStdOut)
	viper.SetDefault(flagLogFileMaxSize, defMaxLogFileSize)
	viper.SetDefault(flagLogFileMaxAge, defMaxLogFileAge)
	viper.SetDefault(flagLogFileMaxBackups, defMaxLogFileBackups)
	viper.SetDefault(flagLogLevel, info)
	viper.SetDefault(flagFileServerPort, fileserver.DefPort)
	viper.SetDefault(flagFileServerPath, path.GetDefaultConfigFilePath())
	err := fileserver.InitFsEncryption()
	if err != nil {
		logger.WithError(err).Fatal("failed to create key for encryption")
	}
	encryptPass, err := fileserver.Encrypt(fileserver.DefAdminPass)
	if err != nil {
		setupErr = err
	}
	viper.SetDefault(flagAdminPass, encryptPass)
	encryptUser, err := fileserver.Encrypt(fileserver.DefAdminUser)
	if err != nil {
		setupErr = err
	}
	viper.SetDefault(flagAdminUser, encryptUser)

	startCmd.Flags().AddFlagSet(configFlagSet)
	startCmd.Flags().Int(flagLogFileMaxBackups, viper.GetInt(flagLogFileMaxBackups), "maximum number of log file rotations")
	startCmd.Flags().Int(flagLogFileMaxSize, viper.GetInt(flagLogFileMaxSize), "maximum size of the log file before it's rotated")
	startCmd.Flags().Int(flagLogFileMaxAge, viper.GetInt(flagLogFileMaxAge), "maximum age of the log file before it's rotated")
	startCmd.Flags().Bool(flagLogFileAndStdout, viper.GetBool(flagLogFileAndStdout), "write logs to stdout if log-file is specified?")
	startCmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	startCmd.Flags().String(flagLogFile, viper.GetString(flagLogFile), "log to file, specify the file location")
	startCmd.Flags().Int(flagFileServerPort, viper.GetInt(flagFileServerPort), "port the file server should listen on")
	startCmd.Flags().String(flagFileServerPath, viper.GetString(flagFileServerPath), "directory to store the filese of the server, starting from home")
	startCmd.Flags().String(flagAdminPass, viper.GetString(flagAdminPass), "password for admin user")
	startCmd.Flags().String(flagAdminUser, viper.GetString(flagAdminUser), "username for admin user")

	err = viper.ReadInConfig()
	if !os.IsNotExist(err) {
		encryptPass, err = fileserver.Encrypt(viper.GetString(flagAdminPass))
		if err != nil {
			setupErr = err
		}
		encryptUser, err = fileserver.Encrypt(viper.GetString(flagAdminUser))
		if err != nil {
			setupErr = err
		}
		viper.Set(flagAdminUser, encryptUser)
		viper.Set(flagAdminPass, encryptPass)
		err = viper.WriteConfig()
		if err != nil {
			setupErr = err
		}
	} else if err != nil {
		setupErr = err
	}

	return startCmd
}




