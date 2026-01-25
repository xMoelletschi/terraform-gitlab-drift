package cmd

import (
	"fmt"
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

	fmt.Printf("Scanning for unmanaged GitLab resources...\n")
	fmt.Printf("  GitLab URL: %s\n", gitlabURL)
	fmt.Printf("  Terraform dir: %s\n", terraformDir)
	fmt.Printf("  Create MR: %v\n", createMR)

	client, err := gitlab.NewClient(gitlabToken, gitlabURL)
	if err != nil {
		fmt.Errorf("TODO")
	}

	// Fetch resources from GitLab API
	resources, err := client.FetchAll(ctx)
	if err != nil {
		fmt.Errorf("TODO")
	}
	// parse files
	// check if its in the files
	// TODO:
	// Write to Terraform files
	// Compare and find unmanaged resources
	// Output results
	// Optionally create MR

	return nil
}
