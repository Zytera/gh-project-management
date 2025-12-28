package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Zytera/gh-project-managment/internal/config"
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gh-project-managment",
	Short:   "TODO",
	Long:    `TODO`,
	Version: Version,
}

func Execute() int {

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ctx := context.WithValue(context.Background(), config.ConfigKey{}, cfg)
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
