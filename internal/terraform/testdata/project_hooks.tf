resource "gitlab_project_hook" "my_group_my_project_hooks_slack_com_services_t123" {
  project                 = gitlab_project.my_group_my_project.id
  url                     = "https://hooks.slack.com/services/T123"
  name                    = "Slack notifications"
  description             = "Posts to #dev"
  enable_ssl_verification = true
  push_events             = true
  merge_requests_events   = true
}

resource "gitlab_project_hook" "my_group_my_project_example_com_webhook" {
  project                 = gitlab_project.my_group_my_project.id
  url                     = "https://example.com/webhook"
  enable_ssl_verification = true
  push_events             = false
  tag_push_events         = true
  pipeline_events         = true
}
