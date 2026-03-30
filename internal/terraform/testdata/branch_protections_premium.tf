resource "gitlab_branch_protection" "my_group_my_project_main" {
  project                      = gitlab_project.my_group_my_project.id
  branch                       = "main"
  push_access_level            = "maintainer"
  merge_access_level           = "developer"
  unprotect_access_level       = "developer"
  code_owner_approval_required = true
  allowed_to_push {
    user_id = 5
  }
  allowed_to_push {
    user_id = 10
  }
  allowed_to_merge {
    group_id = 42
  }
}

resource "gitlab_branch_protection" "my_group_my_project_develop" {
  project            = gitlab_project.my_group_my_project.id
  branch             = "develop"
  push_access_level  = "developer"
  merge_access_level = "developer"
  allow_force_push   = true
}
