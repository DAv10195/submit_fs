package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func NewRootCmd(ctx context.Context, args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: submitFileServer,
		Short: submitFileServer,
		SilenceUsage: true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newStartCommand(ctx, args))
	viper.SetEnvPrefix(submitFileServer)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	return rootCmd
}
