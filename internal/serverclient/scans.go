package serverclient

type Scan struct {
	UUID          string `json:"uuid"`
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id,omitempty"`
	ScanID        string `json:"scan_id"`
	Type          string `json:"type"`
	Environment   string `json:"environment"`
	Status        string `json:"status"`
	CurrentStage  string `json:"current_stage,omitempty"`
	SecurityScore int    `json:"security_score,omitempty"`
	AgentID       string `json:"agent_id,omitempty"`
	TriggeredBy   string `json:"triggered_by,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	CompletedAt   string `json:"completed_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CreateScanPayload struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id,omitempty"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	AgentID     string `json:"agent_id,omitempty"`
}

type UpdateScanPayload struct {
	Status       string `json:"status,omitempty"`
	CurrentStage string `json:"current_stage,omitempty"`
	SecurityScore int   `json:"security_score,omitempty"`
}

type ScansData struct {
	Scan  *Scan  `json:"scan,omitempty"`
	Scans []Scan `json:"scans,omitempty"`
	Stats *struct {
		Total    int                    `json:"total"`
		ByStatus []struct{ Status string; Count int } `json:"by_status"`
		ByType   []struct{ Type string; Count int }   `json:"by_type"`
	} `json:"stats,omitempty"`
}

func (c *Client) ListScans(params ...map[string]string) ([]Scan, error) {
	path := "/api/v1/scans"
	if len(params) > 0 {
		first := true
		for k, v := range params[0] {
			if first {
				path += "?" + k + "=" + v
				first = false
			} else {
				path += "&" + k + "=" + v
			}
		}
	}
	var data ScansData
	err := c.Get(path, &data)
	return data.Scans, err
}

func (c *Client) GetScan(uuid string) (*Scan, error) {
	var data ScansData
	err := c.Get("/api/v1/scans/"+uuid, &data)
	return data.Scan, err
}

func (c *Client) CreateScan(payload CreateScanPayload) (*Scan, error) {
	var data ScansData
	err := c.Post("/api/v1/scans", payload, &data)
	return data.Scan, err
}

func (c *Client) UpdateScan(uuid string, payload UpdateScanPayload) (*Scan, error) {
	var data ScansData
	err := c.Patch("/api/v1/scans/"+uuid, payload, &data)
	return data.Scan, err
}

func (c *Client) DeleteScan(uuid string) error {
	return c.Delete("/api/v1/scans/"+uuid, nil)
}

func (c *Client) GetScanStats() (*struct {
	Total    int                    `json:"total"`
	ByStatus []struct{ Status string; Count int } `json:"by_status"`
	ByType   []struct{ Type string; Count int }   `json:"by_type"`
}, error) {
	var data ScansData
	err := c.Get("/api/v1/scans/stats", &data)
	return data.Stats, err
}
