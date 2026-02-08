resource "gitlab_project" "minimal_project" {
  name                       = "Minimal Project"
  path                       = "minimal-project"
  namespace_id               = 7
  issues_enabled             = false
  merge_requests_enabled     = false
  wiki_enabled               = false
  snippets_enabled           = false
  container_registry_enabled = false
  shared_runners_enabled     = false
  group_runners_enabled      = false
  packages_enabled           = false
}

