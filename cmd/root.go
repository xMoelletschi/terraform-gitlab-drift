package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const defaultGitLabURL = "https://gitlab.com"

var version = "dev"

var (
	terraformDir string
	gitlabToken  string
	gitlabURL    string
	gitlabGroup  string
	verbose      bool
	jsonOutput   bool
)

var rootCmd = &cobra.Command{
	Use:     "terraform-gitlab-drift",
	Short:   "Detect GitLab resources not managed by Terraform",
	Version: version,
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
	rootCmd.PersistentFlags().StringVar(&gitlabURL, "gitlab-url", defaultGitLabURL, "GitLab instance URL")
	rootCmd.PersistentFlags().StringVar(&gitlabGroup, "group", "", "GitLab top-level group to scan (required for gitlab.com)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose (debug) logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output logs in JSON format (useful for CI)")
}

func initLogger() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler = slog.NewTextHandler(os.Stderr, opts)
	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
