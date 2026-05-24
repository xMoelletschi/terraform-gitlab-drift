resource "gitlab_project_job_token_scopes" "my_group_my_project" {
  project = gitlab_project.my_group_my_project.id
  enabled = true
  target_project_ids = [
    gitlab_project.my_group_foo.id,
    999, # unmanaged: other-org/external
  ]
  target_group_ids = [
    888, # unmanaged: other-org/external-group
  ]
}
