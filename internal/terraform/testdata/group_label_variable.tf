variable "gitlab_group_label" {
  description = "Labels for gitlab groups."
  default = {
    "my-group" = {
      "bug" = {
        color       = "#FF0000"
        description = "Bug reports"
      }
      "feature" = {
        color       = "#00FF00"
        description = "Feature requests"
      }
    }
  }
}
