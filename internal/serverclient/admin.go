package serverclient

type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	Role        string `json:"role,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type UsersData struct {
	User  *User  `json:"user,omitempty"`
	Users []User `json:"users,omitempty"`
}

type UpdateUserPayload struct {
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

func (c *Client) ListUsers() ([]User, error) {
	var data UsersData
	err := c.Get("/api/v1/admin/users", &data)
	return data.Users, err
}

func (c *Client) GetUser(uuid string) (*User, error) {
	var data UsersData
	err := c.Get("/api/v1/admin/users/"+uuid, &data)
	return data.User, err
}

func (c *Client) UpdateUser(uuid string, payload UpdateUserPayload) (*User, error) {
	var data UsersData
	err := c.Patch("/api/v1/admin/users/"+uuid, payload, &data)
	return data.User, err
}
