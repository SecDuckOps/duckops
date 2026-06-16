package agent

import "time"

type AgentState string

const (
	StateUnregistered AgentState = "unregistered"
	StateRegistered   AgentState = "registered"
	StateConnected    AgentState = "connected"
	StateDisconnected AgentState = "disconnected"
	StateError        AgentState = "error"
)

type AgentInfo struct {
	ID        string     `json:"id,omitempty"`
	Hostname  string     `json:"hostname"`
	Version   string     `json:"version"`
	Platform  string     `json:"platform"`
	State     AgentState `json:"state"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
}

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
}

type RegisterResponse struct {
	Data struct {
		Agent struct {
			UUID      string `json:"uuid"`
			AgentID   string `json:"agent_id"`
			Status    string `json:"status"`
			Token     string `json:"token,omitempty"`
			CreatedAt string `json:"created_at"`
		} `json:"agent"`
	} `json:"data"`
}

type ConnectRequest struct {
	AgentID string `json:"agent_id"`
}

type ConnectResponse struct {
	Data struct {
		Session struct {
			Token     string `json:"token"`
			ExpiresAt string `json:"expires_at"`
		} `json:"session"`
		Agent struct {
			UUID     string `json:"uuid"`
			Status   string `json:"status"`
			LastSeen string `json:"last_seen"`
		} `json:"agent"`
	} `json:"data"`
}

type HeartbeatRequest struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

type HeartbeatResponse struct {
	Data struct {
		ReceivedAt         string `json:"received_at"`
		NextHeartbeatIn    int    `json:"next_heartbeat_in"`
		PendingJobsCount   int    `json:"pending_jobs_count"`
	} `json:"data"`
}
