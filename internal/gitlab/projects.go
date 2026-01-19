package gitlab

type Project struct {
	ID       int
	Name     string
	Path     string
	FullPath string
}

func (c *Client) ListProjects() ([]Project, error) {}
