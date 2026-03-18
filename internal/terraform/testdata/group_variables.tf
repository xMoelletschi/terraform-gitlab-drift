resource "gitlab_group_variable" "my_group_api_url" {
  group         = gitlab_group.my_group.id
  key           = "API_URL"
  value         = "https://api.example.com"
  variable_type = "env_var"
  protected     = false
  raw           = true
  description   = "API base URL"
}

resource "gitlab_group_variable" "my_group_db_host_production" {
  group             = gitlab_group.my_group.id
  key               = "DB_HOST"
  value             = "db.example.com"
  variable_type     = "env_var"
  protected         = true
  raw               = false
  environment_scope = "production"
}
