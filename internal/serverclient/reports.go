package serverclient

type Report struct {
	UUID          string `json:"uuid"`
	WorkspaceID   string `json:"workspace_id"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	GeneratedBy   string `json:"generated_by,omitempty"`
	GeneratedAt   string `json:"generated_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CreateReportPayload struct {
	WorkspaceID string `json:"workspace_id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
}

type ReportsData struct {
	Report  *Report  `json:"report,omitempty"`
	Reports []Report `json:"reports,omitempty"`
}

func (c *Client) ListReports(workspaceID ...string) ([]Report, error) {
	path := "/api/v1/reports"
	if len(workspaceID) > 0 && workspaceID[0] != "" {
		path += "?workspace_id=" + workspaceID[0]
	}
	var data ReportsData
	err := c.Get(path, &data)
	return data.Reports, err
}

func (c *Client) GetReport(uuid string) (*Report, error) {
	var data ReportsData
	err := c.Get("/api/v1/reports/"+uuid, &data)
	return data.Report, err
}

func (c *Client) CreateReport(payload CreateReportPayload) (*Report, error) {
	var data ReportsData
	err := c.Post("/api/v1/reports", payload, &data)
	return data.Report, err
}

func (c *Client) UpdateReport(uuid string, payload map[string]interface{}) (*Report, error) {
	var data ReportsData
	err := c.Patch("/api/v1/reports/"+uuid, payload, &data)
	return data.Report, err
}
