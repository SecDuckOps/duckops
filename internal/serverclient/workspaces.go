package serverclient

type Workspace struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateWorkspacePayload struct {
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

type WorkspacesData struct {
	Workspace *Workspace  `json:"workspace,omitempty"`
	Workspaces []Workspace `json:"workspaces,omitempty"`
}

func (c *Client) ListWorkspaces() ([]Workspace, error) {
	var data WorkspacesData
	err := c.Get("/api/v1/workspaces", &data)
	return data.Workspaces, err
}

func (c *Client) GetWorkspace(uuid string) (*Workspace, error) {
	var data WorkspacesData
	err := c.Get("/api/v1/workspaces/"+uuid, &data)
	return data.Workspace, err
}

func (c *Client) CreateWorkspace(payload CreateWorkspacePayload) (*Workspace, error) {
	var data WorkspacesData
	err := c.Post("/api/v1/workspaces", payload, &data)
	return data.Workspace, err
}

func (c *Client) UpdateWorkspace(uuid string, payload map[string]interface{}) (*Workspace, error) {
	var data WorkspacesData
	err := c.Patch("/api/v1/workspaces/"+uuid, payload, &data)
	return data.Workspace, err
}

func (c *Client) DeleteWorkspace(uuid string) error {
	return c.Delete("/api/v1/workspaces/"+uuid, nil)
}
