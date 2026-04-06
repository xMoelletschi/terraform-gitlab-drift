resource "gitlab_branch_protection" "my_group_my_project_main" {
  project                = gitlab_project.my_group_my_project.id
  branch                 = "main"
  push_access_level      = "maintainer"
  merge_access_level     = "developer"
  unprotect_access_level = "maintainer"
}

resource "gitlab_branch_protection" "my_group_my_project_release_wildcard" {
  project                = gitlab_project.my_group_my_project.id
  branch                 = "release/*"
  push_access_level      = "no one"
  merge_access_level     = "maintainer"
  unprotect_access_level = "maintainer"
  allow_force_push       = true
}
