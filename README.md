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
| `--include`       | -                    | -                    | Resource types to opt back in (comma-separated). Use to enable resources skipped by default (currently `branch_protection`) |
| `--create-mr`     | -                    | `false`              | Create a merge request with generated Terraform code |
| `--target-repo`   | -                    | *(auto-detected)*    | GitLab project path or ID for the MR              |
| `--mr-branch`     | -                    | `drift/backtrack`    | Branch name for the drift MR                      |
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
├── group_labels.tf         # generated: variable with group → labels
├── project_labels.tf       # generated: variable with project → labels
├── ci_variables.tf         # generated: group and project CI/CD variables
├── pipeline_schedules.tf   # generated: variable with project → pipeline schedules
├── hooks.tf                # generated: project and group webhooks
├── branch_protections.tf   # generated: project branch protection rules
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
- ✅ GitLab Group Labels ([`gitlab_group_label`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/group_label))
- ✅ GitLab Project Labels ([`gitlab_project_label`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project_label))
- ✅ GitLab Pipeline Schedules ([`gitlab_pipeline_schedule`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/pipeline_schedule))
- ✅ GitLab Pipeline Schedule Variables ([`gitlab_pipeline_schedule_variable`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/pipeline_schedule_variable))
- ✅ GitLab Group Variables ([`gitlab_group_variable`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/group_variable)) *
- ✅ GitLab Project Variables ([`gitlab_project_variable`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project_variable)) *
- ✅ GitLab Project Hooks ([`gitlab_project_hook`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project_hook))
- ✅ GitLab Group Hooks ([`gitlab_group_hook`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/group_hook)) *(requires Premium/Ultimate)*
- 🚧 GitLab Branch Protection ([`gitlab_branch_protection`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/branch_protection)) — *skipped by default, opt in with `--include branch_protection`* **
- ✅ GitLab Project Job Token Scopes ([`gitlab_project_job_token_scopes`](https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project_job_token_scopes))
- 🚧 More resources coming soon

> **\* CI/CD Variable Filtering:** Masked variables and file-type variables are automatically skipped.
> Masked variables are excluded because the GitLab API returns redacted values (`[MASKED]`), which would produce invalid Terraform state.
> File-type variables are excluded because they typically contain sensitive data such as SSH private keys or certificates.
>
> **Important:** If you store secrets (SSH keys, API tokens, etc.) as `env_var`-type CI/CD variables, they will be written to `.tf` files in plaintext.
> To prevent this, store sensitive values as **file-type** variables — this is also [GitLab's recommended approach](https://docs.gitlab.com/ci/variables/#use-file-type-cicd-variables) for multi-line secrets like SSH keys, since they cannot be masked.
>
> **\*\* Branch Protection (skipped by default):** The `gitlab_branch_protection` resource is currently skipped by default because of an upstream provider bug —
> the provider's `Read` function does not populate `unprotect_access_level` into Terraform state on Free tier, which causes every `terraform plan` to mark every
> protected branch for forced replacement. Enable it explicitly with `--include branch_protection` once your instance is Premium/Ultimate or once the upstream
> provider bug is fixed.
>
> When enabled, the `allowed_to_push`, `allowed_to_merge`, `allowed_to_unprotect`, `unprotect_access_level`, and `code_owner_approval_required` attributes
> require a GitLab Premium/Ultimate instance. They are emitted by default but can be excluded with `--skip premium`. When skipped, only the free-tier
> attributes (`push_access_level`, `merge_access_level`, `allow_force_push`) are generated.

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
