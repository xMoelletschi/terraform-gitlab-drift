resource "gitlab_group_hook" "my_group_example_com_webhook" {
  group                   = gitlab_group.my_group.id
  url                     = "https://example.com/webhook"
  enable_ssl_verification = true
  push_events             = true
  subgroup_events         = true
}

resource "gitlab_group_hook" "my_group_ci_internal_io_notify" {
  group                   = gitlab_group.my_group.id
  url                     = "https://ci.internal.io/notify"
  name                    = "CI notify"
  enable_ssl_verification = false
  push_events             = true
  merge_requests_events   = true
  emoji_events            = true
  feature_flag_events     = true
}
