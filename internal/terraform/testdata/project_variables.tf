resource "gitlab_project_variable" "my_group_my_project_secret_key" {
  project       = gitlab_project.my_group_my_project.id
  key           = "SECRET_KEY"
  value         = "abc123"
  variable_type = "env_var"
  protected     = false
  raw           = false
  description   = "Secret key"
}
