package serverclient

type Project struct {
	UUID             string `json:"uuid"`
	WorkspaceID      string `json:"workspace_id"`
	Name             string `json:"name"`
	Slug             string `json:"slug,omitempty"`
	RepoURL          string `json:"repo_url,omitempty"`
	Language         string `json:"language,omitempty"`
	HealthScore      int    `json:"health_score,omitempty"`
	SecurityRisk     string `json:"security_risk,omitempty"`
	AutomationMode   string `json:"automation_mode,omitempty"`
	AIAutoFixEnabled bool   `json:"ai_auto_fix_enabled"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type CreateProjectPayload struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	RepoURL     string `json:"repo_url,omitempty"`
	Language    string `json:"language,omitempty"`
}

type ProjectsData struct {
	Project  *Project  `json:"project,omitempty"`
	Projects []Project `json:"projects,omitempty"`
}

func (c *Client) ListProjects(workspaceID ...string) ([]Project, error) {
	path := "/api/v1/projects"
	if len(workspaceID) > 0 && workspaceID[0] != "" {
		path += "?workspace_id=" + workspaceID[0]
	}
	var data ProjectsData
	err := c.Get(path, &data)
	return data.Projects, err
}

func (c *Client) GetProject(uuid string) (*Project, error) {
	var data ProjectsData
	err := c.Get("/api/v1/projects/"+uuid, &data)
	return data.Project, err
}

func (c *Client) CreateProject(payload CreateProjectPayload) (*Project, error) {
	var data ProjectsData
	err := c.Post("/api/v1/projects", payload, &data)
	return data.Project, err
}

func (c *Client) UpdateProject(uuid string, payload map[string]interface{}) (*Project, error) {
	var data ProjectsData
	err := c.Patch("/api/v1/projects/"+uuid, payload, &data)
	return data.Project, err
}

func (c *Client) DeleteProject(uuid string) error {
	return c.Delete("/api/v1/projects/"+uuid, nil)
}
