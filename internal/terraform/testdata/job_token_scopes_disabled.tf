resource "gitlab_project_job_token_scopes" "my_group_my_project" {
  project            = gitlab_project.my_group_my_project.id
  enabled            = false
  target_project_ids = []
}
