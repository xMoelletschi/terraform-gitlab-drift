package gitlab

type Group struct {
	ID       int
	Name     string
	Path     string
	FullPath string
}

func (c *Client) ListGroups() ([]Group, error) {}
