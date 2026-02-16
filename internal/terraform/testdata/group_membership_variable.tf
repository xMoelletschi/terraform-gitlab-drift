variable "gitlab_group_membership" {
  description = "Assign gitlab users to groups."
  default = {
    "my-group" = {
      "jdoe"   = "developer"
      "asmith" = "maintainer"
    }
    "my-group/sub-group" = {}
  }
}
