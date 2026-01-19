package gitlab

type User struct {
	ID    int
	Name  string
	Email string
}

func (c *Client) ListUsers() ([]User, error) {}
