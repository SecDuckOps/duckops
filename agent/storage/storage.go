package storage

import "context"

type StoredIdentity struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type Storage interface {
	LoadIdentity(ctx context.Context) (*StoredIdentity, error)
	SaveIdentity(ctx context.Context, id *StoredIdentity) error
	ClearIdentity(ctx context.Context) error
}
