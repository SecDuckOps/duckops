package serverclient

type EventPayload struct {
	Event       string                 `json:"event"`
	Category    string                 `json:"category"`
	Resource    string                 `json:"resource,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Message     string                 `json:"message,omitempty"`
	AgentID     string                 `json:"agent_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ScanID      string                 `json:"scan_id,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ActivityEvent struct {
	UUID        string                 `json:"uuid"`
	Event       string                 `json:"event"`
	Category    string                 `json:"category"`
	Resource    string                 `json:"resource,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Message     string                 `json:"message,omitempty"`
	AgentID     string                 `json:"agent_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ScanID      string                 `json:"scan_id,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at"`
}

type EventsData struct {
	Events []ActivityEvent `json:"events,omitempty"`
	Event  *ActivityEvent  `json:"event,omitempty"`
}

func (c *Client) CreateEvent(payload EventPayload) (*ActivityEvent, error) {
	var data EventsData
	err := c.Post("/api/v1/events", payload, &data)
	return data.Event, err
}

func (c *Client) ListEvents(params ...map[string]string) ([]ActivityEvent, error) {
	path := "/api/v1/events"
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
	var data EventsData
	err := c.Get(path, &data)
	return data.Events, err
}

func (c *Client) GetEvent(uuid string) (*ActivityEvent, error) {
	var data EventsData
	err := c.Get("/api/v1/events/"+uuid, &data)
	return data.Event, err
}
