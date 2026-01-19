package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	terraformDir string
	gitlabToken  string
	gitlabURL    string
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
	rootCmd.PersistentFlags().StringVar(&terraformDir, "terraform-dir", ".", "Path to Terraform directory")
	rootCmd.PersistentFlags().StringVar(&gitlabToken, "gitlab-token", "", "GitLab API token (or set GITLAB_TOKEN env var)")
	rootCmd.PersistentFlags().StringVar(&gitlabURL, "gitlab-url", "https://gitlab.com", "GitLab instance URL")
}
