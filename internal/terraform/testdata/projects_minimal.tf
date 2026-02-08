resource "gitlab_project" "minimal_project" {
  name                   = "Minimal Project"
  path                   = "minimal-project"
  namespace_id           = 7
  shared_runners_enabled = false
  group_runners_enabled  = false
  packages_enabled       = false
}

