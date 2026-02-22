# terraform-gitlab-drift

[![CI](https://github.com/xMoelletschi/terraform-gitlab-drift/actions/workflows/ci.yml/badge.svg)](https://github.com/xMoelletschi/terraform-gitlab-drift/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/xMoelletschi/terraform-gitlab-drift)](https://github.com/xMoelletschi/terraform-gitlab-drift/releases)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Detect GitLab resources not managed by Terraform and generate Terraform code to bring them under management.

Uses the [GitLab Terraform Provider](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs) resource definitions.

## Features

- 🔍 **Drift Detection**: Scan GitLab groups and projects to identify resources not managed by Terraform
- 📝 **Code Generation**: Automatically generate Terraform code for unmanaged resources
- 🔄 **Diff Comparison**: Show differences between existing and generated Terraform configurations
- 📦 **Import Commands**: Generate `terraform import` commands for new resources
- 🔀 **Merge Request Creation**: Automatically create or update GitLab MRs with generated `.tf` files
- 🐳 **Docker-ready**: Designed for CI/CD pipelines

## Quick Start

### Local Installation

```bash
go install github.com/xMoelletschi/terraform-gitlab-drift@latest
terraform-gitlab-drift scan --group my-group
```

## GitLab CI Usage

### Basic Drift Check

```yaml
drift-check:
  image: ghcr.io/xmoelletschi/terraform-gitlab-drift:latest
  script:
    - terraform-gitlab-drift scan --group $CI_PROJECT_ROOT_NAMESPACE
```

### Auto-create MR on Drift

```yaml
drift-remediation:
  image: ghcr.io/xmoelletschi/terraform-gitlab-drift:latest
  script:
    - terraform-gitlab-drift scan --group $CI_PROJECT_ROOT_NAMESPACE --create-mr --show-diff=false
```

When drift is detected, this creates (or updates) a merge request in the current repository with the generated `.tf` files. The target repo is auto-detected from the git remote; use `--target-repo` to override.

## Configuration

### Command-line Flags

| Flag              | Environment Variable | Default              | Description                                       |
| ----------------- | -------------------- | -------------------- | ------------------------------------------------- |
| `--gitlab-token`  | `GITLAB_TOKEN`       | -                    | GitLab API token (required)                       |
| `--gitlab-url`    | -                    | `https://gitlab.com` | GitLab instance URL                               |
| `--group`         | -                    | -                    | Top-level group to scan (required for gitlab.com) |
| `--terraform-dir` | -                    | `.`                  | Path to Terraform directory                       |
| `--overwrite`     | -                    | `false`              | Overwrite files in terraform directory            |
| `--show-diff`     | -                    | `true`               | Show diff between generated and existing files    |
| `--skip`          | -                    | -                    | Resource types to skip (comma-separated). Use `premium` to skip all Premium-tier resources |
| `--create-mr`     | -                    | `false`              | Create a merge request with generated Terraform code |
| `--target-repo`   | -                    | *(auto-detected)*    | GitLab project path or ID for the MR              |
| `--mr-dest-path`  | -                    | *(root)*             | Path within target repo where `.tf` files go      |
| `--verbose`, `-v` | -                    | `false`              | Enable verbose (debug) logging                    |
| `--json`          | -                    | `false`              | Output logs in JSON format                        |

### Directory Structure

The tool generates one `.tf` file per GitLab namespace, using normalized names (lowercase, `/` and `-` replaced with `_`). Your Terraform directory should follow this structure to get accurate drift detection:

```
terraform/
├── backend.tf
├── providers.tf
├── my_group.tf             # generated: top-level group + its projects
├── my_group_sub_group.tf   # generated: sub-group + its projects
├── group_membership.tf     # generated: variable with group → user memberships
├── project_membership.tf   # generated: variable with project → shared groups
└── ...
```

> **Important:** The drift check only compares files that match the generated filenames.
> If you have resources defined in differently named files (e.g. `main.tf`, `projects.tf`),
> they will not be detected and the tool will report those resources as unmanaged.
>
> To fix this, move your resource definitions into the files matching the generated naming
> convention, or use `--overwrite` to let the tool manage the file structure for you.

### Supported Resources

- ✅ GitLab Groups ([`gitlab_group`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/group))
- ✅ GitLab Group Memberships ([`gitlab_group_membership`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/group_membership))
- ✅ GitLab Projects ([`gitlab_project`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project))
- ✅ GitLab Project Share Groups ([`gitlab_project_share_group`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project_share_group))
- 🚧 More resources coming soon

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Push to the branch (`git push origin feature/amazing-feature`)
4. Open a Pull Request

Please make sure to:

- Add tests for new features
- Update documentation as needed
- Ensure CI checks pass

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [GitLab Go SDK](https://gitlab.com/gitlab-org/api/client-go) - GitLab API client
- [HCL](https://github.com/hashicorp/hcl) - Terraform configuration parsing

---

**Note**: This tool is not affiliated with HashiCorp or GitLab.
