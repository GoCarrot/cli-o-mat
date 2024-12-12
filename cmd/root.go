package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/config"
)

// nolint: gochecknoglobals
var rootCmd = &cobra.Command{
	Use:   "cli-o-mat",
	Short: "CLI tool for managing Omat deploys",
	Long: `cli-o-mat is a tool for seeing what's deployable, what's deployed,
initiating deploys, and cancelling deploys using Teak.io's Omat infrastructure
tooling.`,
}

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}

// nolint: gochecknoinits
func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "", "", "Which AWS region to operate in")
}

// nolint: gochecknoglobals
var (
	region string
)

func loadOmatConfig(accountName string) *config.Omat {
	omat := config.NewOmat(accountName)

	omat.LoadConfig()

	if region != "" {
		omat.Region = region
	}

	return omat
}
