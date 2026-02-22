package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/terraform"
)

var (
	createMR      bool
	overwrite     bool
	showDiff      bool
	skipResources []string
	targetRepo    string
	mrDestPath    string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for GitLab resources not managed by Terraform",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&createMR, "create-mr", false, "Create a merge request with generated Terraform code")
	scanCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite files in terraform directory (default: write to tmp/ subdirectory)")
	scanCmd.Flags().BoolVar(&showDiff, "show-diff", true, "Show diff between generated and existing files")
	scanCmd.Flags().StringSliceVar(&skipResources, "skip", nil, "Resource types to skip (comma-separated). Use 'premium' to skip all Premium-tier resources")
	scanCmd.Flags().StringVar(&targetRepo, "target-repo", "", "GitLab project path or ID for the MR (default: detected from git remote in --terraform-dir)")
	scanCmd.Flags().StringVar(&mrDestPath, "mr-dest-path", "", "Path within target repo where .tf files go (default: root)")
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

	if createMR && targetRepo == "" {
		detected, err := detectGitLabProject(terraformDir, gitlabURL)
		if err != nil {
			return fmt.Errorf("--target-repo not set and could not detect from git remote: %w", err)
		}
		targetRepo = detected
		slog.Info("detected target repo from git remote", "target_repo", targetRepo)
	}

	skipSet, skipWarnings := skip.Parse(skipResources)
	for _, w := range skipWarnings {
		slog.Warn("unknown skip value, ignoring", "name", w)
	}
	if len(skipSet) > 0 {
		skipped := make([]string, 0, len(skipSet))
		for k := range skipSet {
			skipped = append(skipped, k)
		}
		slog.Info("skipping resource types", "skipped", skipped)
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
	resources, err := client.FetchAll(ctx, skipSet)
	if err != nil {
		return fmt.Errorf("fetching resources: %w", err)
	}

	groupMemberCount := 0
	for _, members := range resources.GroupMembers {
		groupMemberCount += len(members)
	}

	projectShareGroupCount := 0
	for _, p := range resources.Projects {
		projectShareGroupCount += len(p.SharedWithGroups)
	}

	slog.Info("fetched resources",
		"groups", len(resources.Groups),
		"projects", len(resources.Projects),
		"group_members", groupMemberCount,
		"project_share_groups", projectShareGroupCount,
	)

	outputDir := filepath.Join(terraformDir, "tmp")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating tmp directory: %w", err)
	}

	if err := terraform.WriteAll(resources, outputDir, gitlabGroup, skipSet); err != nil {
		return fmt.Errorf("writing terraform files: %w", err)
	}

	slog.Info("wrote terraform files", "dir", outputDir)

	// Compare generated .tf files with existing ones
	driftFound := false
	files, err := filepath.Glob(filepath.Join(outputDir, "*.tf"))
	if err != nil {
		return fmt.Errorf("listing generated files: %w", err)
	}

	for _, genFile := range files {
		base := filepath.Base(genFile)
		existingFile := filepath.Join(terraformDir, base)

		if _, err := os.Stat(existingFile); os.IsNotExist(err) {
			slog.Warn("new unmanaged resource detected", "file", base)
			if showDiff {
				diffCmd := exec.Command("diff", "-u", "--color=auto", "/dev/null", genFile)
				diffCmd.Stdout = os.Stdout
				diffCmd.Stderr = os.Stderr
				_ = diffCmd.Run()
			}
			driftFound = true
			continue
		}

		if showDiff {
			diffCmd := exec.Command("diff", "-u", "--color=auto", existingFile, genFile)
			diffCmd.Stdout = os.Stdout
			diffCmd.Stderr = os.Stderr
			if err := diffCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
					driftFound = true
					continue
				}
				return fmt.Errorf("diff command failed for %s: %w", base, err)
			}
		} else {
			genData, err := os.ReadFile(genFile)
			if err != nil {
				return fmt.Errorf("reading generated file %s: %w", base, err)
			}
			existingData, err := os.ReadFile(existingFile)
			if err != nil {
				return fmt.Errorf("reading existing file %s: %w", base, err)
			}
			if !bytes.Equal(genData, existingData) {
				driftFound = true
			}
		}
	}

	// Generate import commands for new resources
	existingResources, err := terraform.ParseExistingResources(terraformDir)
	if err != nil {
		return fmt.Errorf("parsing existing terraform files: %w", err)
	}
	importCmds := terraform.GenerateImportCommands(resources, existingResources, gitlabGroup, skipSet)
	if len(importCmds) > 0 {
		driftFound = true
		if _, err := fmt.Fprintln(os.Stdout, "\nImport commands for new resources:"); err != nil {
			return fmt.Errorf("printing import commands: %w", err)
		}
		if err := terraform.PrintImportCommands(os.Stdout, importCmds); err != nil {
			return fmt.Errorf("printing import commands: %w", err)
		}
	}

	// Create or update a merge request if drift was found
	if createMR {
		if !driftFound {
			slog.Info("no drift detected, skipping MR creation")
		} else {
			result, err := createDriftMR(ctx, client, targetRepo, outputDir, mrDestPath)
			if err != nil {
				return fmt.Errorf("creating drift MR: %w", err)
			}
			if result.Created {
				slog.Info("created merge request", "url", result.WebURL)
			} else {
				slog.Info("updated existing merge request", "url", result.WebURL)
			}
		}
	}

	// If --overwrite, copy generated files from tmp/ to the terraform directory
	if overwrite {
		owFiles, err := filepath.Glob(filepath.Join(outputDir, "*.tf"))
		if err != nil {
			return fmt.Errorf("listing generated files: %w", err)
		}
		for _, src := range owFiles {
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", src, err)
			}
			dst := filepath.Join(terraformDir, filepath.Base(src))
			if err := os.WriteFile(dst, data, 0644); err != nil {
				return fmt.Errorf("writing file %s: %w", dst, err)
			}
		}
		slog.Info("overwrote terraform files", "dir", terraformDir)
	}

	if driftFound {
		slog.Warn("drift detected")
		os.Exit(1)
	}

	slog.Info("no drift detected")

	return nil
}

func createDriftMR(ctx context.Context, client *gitlab.Client, project, outputDir, destPath string) (*gitlab.DriftMRResult, error) {
	defaultBranch, err := client.GetDefaultBranch(ctx, project)
	if err != nil {
		return nil, err
	}

	existingMR, err := client.FindExistingDriftMR(ctx, project)
	if err != nil {
		return nil, err
	}

	branchName := "drift/update-" + time.Now().Format("2006-01-02")
	if existingMR != nil {
		branchName = existingMR.SourceBranch
	}

	if err := client.EnsureBranch(ctx, project, branchName, defaultBranch); err != nil {
		return nil, err
	}

	committed, err := client.CommitDriftFiles(ctx, project, branchName, outputDir, destPath)
	if err != nil {
		return nil, err
	}

	if !committed {
		slog.Info("no file changes to commit, all files are identical")
	}

	if existingMR != nil {
		return &gitlab.DriftMRResult{
			WebURL:  existingMR.WebURL,
			Created: false,
		}, nil
	}

	mr, err := client.CreateDriftMR(ctx, project, branchName, defaultBranch)
	if err != nil {
		return nil, err
	}

	return &gitlab.DriftMRResult{
		WebURL:  mr.WebURL,
		Created: true,
	}, nil
}

// detectGitLabProject extracts the GitLab project path from the git remote
// origin URL of the given directory. It handles both SSH and HTTPS remotes.
// Example: "git@gitlab.com:group/project.git" → "group/project"
// Example: "https://gitlab.com/group/project.git" → "group/project"
func detectGitLabProject(dir, gitlabBaseURL string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running git remote get-url origin: %w", err)
	}
	remote := strings.TrimSpace(string(out))

	baseURL, _ := url.Parse(gitlabBaseURL)

	// SSH format: git@gitlab.com:group/subgroup/project.git
	if strings.HasPrefix(remote, "git@") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("unexpected SSH remote format: %s", remote)
		}
		// "git@gitlab.com" → "gitlab.com"
		sshHost := strings.TrimPrefix(parts[0], "git@")
		if baseURL != nil && sshHost != baseURL.Host {
			return "", fmt.Errorf("git remote host %q does not match --gitlab-url host %q", sshHost, baseURL.Host)
		}
		return strings.TrimSuffix(parts[1], ".git"), nil
	}

	// HTTPS format: https://gitlab.com/group/subgroup/project.git
	u, err := url.Parse(remote)
	if err != nil {
		return "", fmt.Errorf("parsing remote URL %s: %w", remote, err)
	}
	if baseURL != nil && u.Host != baseURL.Host {
		return "", fmt.Errorf("git remote host %q does not match --gitlab-url host %q", u.Host, baseURL.Host)
	}
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	if path == "" {
		return "", fmt.Errorf("could not extract project path from remote URL: %s", remote)
	}

	return path, nil
}
