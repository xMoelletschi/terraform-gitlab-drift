resource "gitlab_project" "my_group_minimal_project" {
  name         = "Minimal Project"
  path         = "minimal-project"
  namespace_id = gitlab_group.my_group.id
}

