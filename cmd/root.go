package cmd

import (
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const (
	defaultGitLabURL = "https://gitlab.com"
	devVersion       = "dev"
	shortSHALen      = 7
)

var version = devVersion

var (
	terraformDir string
	gitlabToken  string
	gitlabURL    string
	gitlabGroup  string
	verbose      bool
	jsonOutput   bool
)

var rootCmd = &cobra.Command{
	Use:   "terraform-gitlab-drift",
	Short: "Detect GitLab resources not managed by Terraform",
}

func resolveVersion() string {
	if version != devVersion {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if revision == "" {
		return version
	}
	short := revision
	if len(short) > shortSHALen {
		short = short[:shortSHALen]
	}
	if modified == "true" {
		return "dev-" + short + "+dirty"
	}
	return "dev-" + short
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.Version = resolveVersion()

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
