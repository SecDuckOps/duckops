package serverclient

type Pipeline struct {
	UUID          string `json:"uuid"`
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	BranchName    string `json:"branch_name,omitempty"`
	CommitHash    string `json:"commit_hash,omitempty"`
	CommitAuthor  string `json:"commit_author,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	TriggeredBy   string `json:"triggered_by,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	CompletedAt   string `json:"completed_at,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type CreatePipelinePayload struct {
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id,omitempty"`
	Name          string `json:"name"`
	BranchName    string `json:"branch_name,omitempty"`
	CommitHash    string `json:"commit_hash,omitempty"`
	CommitAuthor  string `json:"commit_author,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
}

type PipelinesData struct {
	Pipeline  *Pipeline  `json:"pipeline,omitempty"`
	Pipelines []Pipeline `json:"pipelines,omitempty"`
}

func (c *Client) ListPipelines(workspaceID ...string) ([]Pipeline, error) {
	path := "/api/v1/pipelines"
	if len(workspaceID) > 0 && workspaceID[0] != "" {
		path += "?workspace_id=" + workspaceID[0]
	}
	var data PipelinesData
	err := c.Get(path, &data)
	return data.Pipelines, err
}

func (c *Client) GetPipeline(uuid string) (*Pipeline, error) {
	var data PipelinesData
	err := c.Get("/api/v1/pipelines/"+uuid, &data)
	return data.Pipeline, err
}

func (c *Client) CreatePipeline(payload CreatePipelinePayload) (*Pipeline, error) {
	var data PipelinesData
	err := c.Post("/api/v1/pipelines", payload, &data)
	return data.Pipeline, err
}

func (c *Client) UpdatePipeline(uuid string, payload map[string]interface{}) (*Pipeline, error) {
	var data PipelinesData
	err := c.Patch("/api/v1/pipelines/"+uuid, payload, &data)
	return data.Pipeline, err
}

func (c *Client) DeletePipeline(uuid string) error {
	return c.Delete("/api/v1/pipelines/"+uuid, nil)
}
