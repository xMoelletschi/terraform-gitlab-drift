resource "gitlab_group" "parent_group" {
  name = "Parent Group"
  path = "parent-group"
}

resource "gitlab_group" "full_group" {
  name                              = "Full Group"
  path                              = "full-group"
  description                       = "Full group description"
  visibility_level                  = "private"
  parent_id                         = gitlab_group.parent_group.id
  lfs_enabled                       = true
  request_access_enabled            = true
  membership_lock                   = true
  share_with_group_lock             = true
  require_two_factor_authentication = true
  two_factor_grace_period           = 7
  project_creation_level            = "owner"
  subgroup_creation_level           = "maintainer"
  auto_devops_enabled               = true
  emails_enabled                    = false
  mentions_disabled                 = true
  prevent_forking_outside_group     = true
  shared_runners_setting            = "disabled_and_overridable"
  default_branch                    = "main"
  wiki_access_level                 = "private"
  ip_restriction_ranges             = "10.0.0.0/24"
  default_branch_protection_defaults {
    allowed_to_push            = [40]
    allowed_to_merge           = [40]
    allow_force_push           = true
    developer_can_initial_push = true
  }
}

