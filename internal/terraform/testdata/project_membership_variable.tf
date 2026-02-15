variable "gitlab_project_membership" {
  description = "Share projects with groups."
  default = {
    "my-group/project-a" = {
      "my-group/sub-group" = "developer"
    }
    "my-group/project-b" = {}
  }
}
