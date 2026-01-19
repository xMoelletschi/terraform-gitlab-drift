package gitlab

import (
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	api *gl.Client
}

type Resources struct {
	Groups   []Group
	Projects []Project
	Users    []User
}

func NewClient(token, baseURL string) (*Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}
	return &Client{api: client}, nil
}

func (c *Client) FetchAll() (*Resources, error) {
}
