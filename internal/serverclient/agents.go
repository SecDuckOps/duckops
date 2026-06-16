package serverclient

type Agent struct {
	UUID            string                 `json:"uuid"`
	AgentID         string                 `json:"agent_id"`
	Hostname        string                 `json:"hostname"`
	Version         string                 `json:"version"`
	Platform        string                 `json:"platform"`
	Status          string                 `json:"status"`
	Capabilities    []string               `json:"capabilities,omitempty"`
	LastHeartbeatAt string                 `json:"last_heartbeat_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

type RegisterAgentPayload struct {
	AgentID      string   `json:"agent_id"`
	Hostname     string   `json:"hostname"`
	Version      string   `json:"version"`
	Platform     string   `json:"platform"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type HeartbeatPayload struct {
	AgentID string  `json:"agent_id"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu,omitempty"`
	Memory  float64 `json:"memory,omitempty"`
}

type AgentsData struct {
	Agent  *Agent  `json:"agent,omitempty"`
	Agents []Agent `json:"agents,omitempty"`
}

func (c *Client) RegisterAgent(payload RegisterAgentPayload) (*Agent, error) {
	var data AgentsData
	err := c.Post("/api/v1/agents/register", payload, &data)
	return data.Agent, err
}

func (c *Client) Heartbeat(payload HeartbeatPayload) (*Agent, error) {
	var data AgentsData
	err := c.Post("/api/v1/agents/heartbeat", payload, &data)
	return data.Agent, err
}

func (c *Client) Connect() error {
	return c.Post("/api/v1/agents/connect", nil, nil)
}

func (c *Client) ListAgents() ([]Agent, error) {
	var data AgentsData
	err := c.Get("/api/v1/agents", &data)
	return data.Agents, err
}

func (c *Client) GetAgent(uuid string) (*Agent, error) {
	var data AgentsData
	err := c.Get("/api/v1/agents/"+uuid, &data)
	return data.Agent, err
}

func (c *Client) UpdateAgent(uuid string, payload map[string]interface{}) (*Agent, error) {
	var data AgentsData
	err := c.Patch("/api/v1/agents/"+uuid, payload, &data)
	return data.Agent, err
}
