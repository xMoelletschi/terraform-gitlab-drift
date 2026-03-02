resource "gitlab_project_label" "my_group_my_project" {
  for_each    = var.gitlab_project_label["my-group/my-project"]
  project     = gitlab_project.my_group_my_project.id
  name        = each.key
  color       = each.value.color
  description = each.value.description
}
