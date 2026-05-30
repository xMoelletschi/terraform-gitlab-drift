resource "gitlab_tag_protection" "my_group_my_project_v1_0_0" {
  project             = gitlab_project.my_group_my_project.id
  tag                 = "v1.0.0"
  create_access_level = "maintainer"
}

resource "gitlab_tag_protection" "my_group_my_project_release_wildcard" {
  project             = gitlab_project.my_group_my_project.id
  tag                 = "release-*"
  create_access_level = "developer"
}
