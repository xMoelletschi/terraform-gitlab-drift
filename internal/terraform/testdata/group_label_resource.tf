resource "gitlab_group_label" "my_group" {
  for_each    = var.gitlab_group_label["my-group"]
  group       = gitlab_group.my_group.id
  name        = each.key
  color       = each.value.color
  description = each.value.description
}
