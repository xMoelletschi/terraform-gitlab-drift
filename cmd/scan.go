package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
)

var createMR bool

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for GitLab resources not managed by Terraform",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&createMR, "create-mr", false, "Create a merge request with generated Terraform code")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	token := gitlabToken
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitLab token required: use --gitlab-token flag or set GITLAB_TOKEN environment variable")
	}

	if gitlabGroup == "" && gitlabURL == defaultGitLabURL {
		return fmt.Errorf("--group is required when using gitlab.com, specify your top-level group")
	}

	slog.Info("scanning for unmanaged GitLab resources",
		"gitlab_url", gitlabURL,
		"group", gitlabGroup,
		"terraform_dir", terraformDir,
		"create_mr", createMR,
	)

	client, err := gitlab.NewClient(token, gitlabURL, gitlabGroup)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	slog.Debug("fetching resources from GitLab API")

	// Fetch resources from GitLab API
	resources, err := client.FetchAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching resources: %w", err)
	}

	slog.Info("fetched resources",
		"groups", len(resources.Groups),
		"projects", len(resources.Projects),
	)
	// parse files
	// check if its in the files
	// TODO:
	// Write to Terraform files
	// Compare and find unmanaged resources
	// Output results
	// Optionally create MR

	return nil
}
