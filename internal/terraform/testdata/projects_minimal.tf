resource "gitlab_project" "minimal_project" {
  name         = "Minimal Project"
  path         = "minimal-project"
  namespace_id = 7
}

