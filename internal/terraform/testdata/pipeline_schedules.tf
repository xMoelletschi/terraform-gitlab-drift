resource "gitlab_pipeline_schedule" "my_group_my_project_nightly_build" {
  project       = gitlab_project.my_group_my_project.id
  description   = "Nightly build"
  ref           = "main"
  cron          = "0 2 * * *"
  cron_timezone = "Europe/Vienna"
  active        = true
}

resource "gitlab_pipeline_schedule_variable" "my_group_my_project_nightly_build_deploy_env" {
  project              = gitlab_project.my_group_my_project.id
  pipeline_schedule_id = gitlab_pipeline_schedule.my_group_my_project_nightly_build.pipeline_schedule_id
  key                  = "DEPLOY_ENV"
  value                = "staging"
  variable_type        = "env_var"
}

resource "gitlab_pipeline_schedule" "my_group_my_project_weekly_cleanup" {
  project       = gitlab_project.my_group_my_project.id
  description   = "Weekly cleanup"
  ref           = "main"
  cron          = "0 0 * * 0"
  cron_timezone = "UTC"
  active        = false
}
