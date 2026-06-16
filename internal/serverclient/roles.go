package serverclient

type Role struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsSystem    bool   `json:"is_system"`
	CreatedAt   string `json:"created_at"`
}

type CreateRolePayload struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RolesData struct {
	Role  *Role  `json:"role,omitempty"`
	Roles []Role `json:"roles,omitempty"`
}

func (c *Client) ListRoles() ([]Role, error) {
	var data RolesData
	err := c.Get("/api/v1/roles", &data)
	return data.Roles, err
}

func (c *Client) CreateRole(payload CreateRolePayload) (*Role, error) {
	var data RolesData
	err := c.Post("/api/v1/roles", payload, &data)
	return data.Role, err
}

func (c *Client) UpdateRole(uuid string, payload map[string]interface{}) (*Role, error) {
	var data RolesData
	err := c.Patch("/api/v1/roles/"+uuid, payload, &data)
	return data.Role, err
}

type Permission struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

type PermissionsData struct {
	Permissions []Permission `json:"permissions,omitempty"`
}

func (c *Client) ListPermissions() ([]Permission, error) {
	var data PermissionsData
	err := c.Get("/api/v1/permissions", &data)
	return data.Permissions, err
}
