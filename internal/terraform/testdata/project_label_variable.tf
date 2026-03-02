variable "gitlab_project_label" {
  description = "Labels for gitlab projects."
  default = {
    "my-group/my-project" = {
      "urgent" = {
        color       = "#FF0000"
        description = "Urgent issues"
      }
    }
  }
}
