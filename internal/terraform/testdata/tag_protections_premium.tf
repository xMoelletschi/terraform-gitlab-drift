resource "gitlab_tag_protection" "my_group_my_project_v1_0_0" {
  project             = gitlab_project.my_group_my_project.id
  tag                 = "v1.0.0"
  create_access_level = "maintainer"
  allowed_to_create {
    user_id = 5
  }
  allowed_to_create {
    user_id = 10
  }
}

resource "gitlab_tag_protection" "my_group_my_project_release_wildcard" {
  project             = gitlab_project.my_group_my_project.id
  tag                 = "release-*"
  create_access_level = "developer"
  allowed_to_create {
    group_id = 42
  }
}
