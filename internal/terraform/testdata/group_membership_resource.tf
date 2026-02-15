resource "gitlab_group_membership" "my_group" {
  for_each     = var.gitlab_group_membership["my-group"]
  group_id     = gitlab_group.my_group.id
  user_id      = gitlab_user.main[each.key].id
  access_level = each.value
}
