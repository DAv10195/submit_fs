package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/DAv10195/submit_fs/path"
	"github.com/DAv10195/submit_fs/server"
	"github.com/gorilla/mux"
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
	"strings"
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

			if err := server.InitFsEncryption(); err != nil {
				logger.WithError(err).Fatal("error setting up encryption")
			}
			encryptPass := viper.GetString(flagPassword)
			if !strings.Contains(encryptPass,"encrypted:"){
				encryptPass, err = server.Encrypt(encryptPass)
				if err != nil {
					logger.WithError(err).Fatal("error encrypting password")
				}
			}
			viper.Set(flagPassword, strings.TrimPrefix(encryptPass, "encrypted:"))
			usedConfFile := viper.ConfigFileUsed()
			if usedConfFile != "" {
				if _, err := os.Stat(usedConfFile); err != nil {
					if !os.IsNotExist(err) {
						logger.WithError(err).Fatalf("error accessing config file at %s", usedConfFile)
					}
				} else {
					// replace the password section with the encrypted password
					if !strings.Contains(encryptPass, "encrypted:"){
						encryptConfig(encryptPass)
					}
				}
			}
			baseRouter := mux.NewRouter()
			baseRouter.Use(server.AuthenticationMiddleware)
			filesPath := filepath.Join(viper.GetString(flagFileServerPath), "files")
			if err := os.MkdirAll(filesPath, 0700); err != nil {
				logrus.WithError(err).Fatalf("error creating \"files\" directory for storing files in %s", viper.GetString(flagFileServerPath))
			}
			router := server.InitRouters(baseRouter, filesPath)
			server.InitFolders()
			fs := server.NewFileServer(router)
			go func() {
				if err := fs.ListenAndServe(); err != http.ErrServerClosed {
					logger.WithError(err).Fatal("submit fs crashed")
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
	_ = configFlagSet.StringP(flagConfigFile, "c", "", "path to submit fs config file")
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
	viper.SetDefault(flagFileServerPort, server.DefPort)
	viper.SetDefault(flagFileServerPath, path.GetDefaultWorkDirPath())
	viper.SetDefault(flagPassword, server.DefPass)
	viper.SetDefault(flagUser, server.DefUser)

	startCmd.Flags().AddFlagSet(configFlagSet)
	startCmd.Flags().Int(flagLogFileMaxBackups, viper.GetInt(flagLogFileMaxBackups), "maximum number of log file rotations")
	startCmd.Flags().Int(flagLogFileMaxSize, viper.GetInt(flagLogFileMaxSize), "maximum size of the log file before it's rotated")
	startCmd.Flags().Int(flagLogFileMaxAge, viper.GetInt(flagLogFileMaxAge), "maximum age of the log file before it's rotated")
	startCmd.Flags().Bool(flagLogFileAndStdout, viper.GetBool(flagLogFileAndStdout), "write logs to stdout if log-file is specified?")
	startCmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	startCmd.Flags().String(flagLogFile, viper.GetString(flagLogFile), "log to file, specify the file location")
	startCmd.Flags().Int(flagFileServerPort, viper.GetInt(flagFileServerPort), "port the file server should listen on")
	startCmd.Flags().String(flagFileServerPath, viper.GetString(flagFileServerPath), "directory to store the files of the server, starting from home")
	startCmd.Flags().String(flagPassword, viper.GetString(flagPassword), "password")
	startCmd.Flags().String(flagUser, viper.GetString(flagUser), "username")

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		setupErr = err
	}

	return startCmd
}


func encryptConfig(encryptPass string) error{
	var s = make([]string,2)
	s[0] = "password:"
	s[1] = "encrypted:" + encryptPass
	confPath := filepath.Join(viper.GetString(flagFileServerPath),defaultConfigFileName)
	confLines, err := readLines(confPath)
	if err != nil {
		return err
	}
	for i:=0; i<len(confLines);i++{
		if strings.Contains(confLines[i], "password:"){
			confLines[i] = strings.Join(s," ")
		}
	}
	err = writeLines(confLines, confPath)
	if err != nil {
		return err
	}
	return nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}