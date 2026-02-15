resource "gitlab_project_share_group" "my_group_project_a" {
  for_each     = var.gitlab_project_membership["my-group/project-a"]
  project      = gitlab_project.my_group_project_a.id
  group_id     = data.gitlab_group.by_projects[each.key].id
  group_access = each.value
}
