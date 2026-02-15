locals {
  groups_by_projects = toset(distinct(flatten([
    for key, project in var.gitlab_project_membership : [
      for group, permission in project : [
        group
      ]
    ]
  ])))
}

data "gitlab_group" "by_projects" {
  for_each  = local.groups_by_projects
  full_path = each.key
}
