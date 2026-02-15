package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/terraform"
)

var (
	createMR  bool
	overwrite bool
	showDiff  bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for GitLab resources not managed by Terraform",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&createMR, "create-mr", false, "[TODO:WIP] Create a merge request with generated Terraform code")
	scanCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite files in terraform directory (default: write to tmp/ subdirectory)")
	scanCmd.Flags().BoolVar(&showDiff, "show-diff", true, "Show diff between generated and existing files")
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

	outputDir := terraformDir
	if !overwrite {
		outputDir = filepath.Join(terraformDir, "tmp")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("creating tmp directory: %w", err)
		}
	}

	if err := terraform.WriteAll(resources, outputDir, gitlabGroup); err != nil {
		return fmt.Errorf("writing terraform files: %w", err)
	}

	slog.Info("wrote terraform files", "dir", outputDir, "overwrite", overwrite)

	// Compare generated .tf files with existing ones using diff command
	if showDiff {
		slog.Info("comparing files", "existing", terraformDir, "generated", outputDir)

		files, err := filepath.Glob(filepath.Join(outputDir, "*.tf"))
		if err != nil {
			return fmt.Errorf("listing generated files: %w", err)
		}

		driftFound := false
		for _, genFile := range files {
			base := filepath.Base(genFile)
			existingFile := filepath.Join(terraformDir, base)

			if _, err := os.Stat(existingFile); os.IsNotExist(err) {
				slog.Warn("new unmanaged resource detected", "file", base)
				diffCmd := exec.Command("diff", "-u", "--color=auto", "/dev/null", genFile)
				diffCmd.Stdout = os.Stdout
				diffCmd.Stderr = os.Stderr
				_ = diffCmd.Run()
				driftFound = true
				continue
			}

			diffCmd := exec.Command("diff", "-u", "--color=auto", existingFile, genFile)
			diffCmd.Stdout = os.Stdout
			diffCmd.Stderr = os.Stderr

			if err := diffCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() == 1 {
						driftFound = true
						continue
					}
					return fmt.Errorf("diff command failed for %s: %w", base, err)
				}
				return fmt.Errorf("running diff command for %s: %w", base, err)
			}
		}

		if driftFound {
			slog.Warn("drift detected")
			os.Exit(1)
		}

		slog.Info("no drift detected")
	}

	return nil
}
