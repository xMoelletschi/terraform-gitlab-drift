package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	terraformDir string
	gitlabToken  string
	gitlabURL    string
	verbose      bool
	jsonOutput   bool
)

var rootCmd = &cobra.Command{
	Use:   "terraform-gitlab-drift",
	Short: "Detect GitLab resources not managed by Terraform",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().StringVar(&terraformDir, "terraform-dir", ".", "Path to Terraform directory")
	rootCmd.PersistentFlags().StringVar(&gitlabToken, "gitlab-token", "", "GitLab API token (or set GITLAB_TOKEN env var)")
	rootCmd.PersistentFlags().StringVar(&gitlabURL, "gitlab-url", "https://gitlab.com", "GitLab instance URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose (debug) logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output logs in JSON format (useful for CI)")
}

func initLogger() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
