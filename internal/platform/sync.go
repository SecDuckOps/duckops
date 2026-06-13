package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Syncer struct {
	client *Client
	done   chan struct{}
	wg     sync.WaitGroup

	hostname   string
	version    string
	capabilities []string

	mu        sync.RWMutex
	status    string
	replaying bool
}

func NewSyncer(client *Client, hostname, version string) *Syncer {
	return &Syncer{
		client:   client,
		done:     make(chan struct{}),
		hostname: hostname,
		version:  version,
		status:   "initialized",
	}
}

func (s *Syncer) Start(ctx context.Context) {
	s.wg.Add(3)
	go s.agentRegistrationLoop(ctx)
	go s.heartbeatLoop(ctx)
	go s.queueReplayLoop(ctx)
}

func (s *Syncer) Stop() {
	close(s.done)
	s.wg.Wait()
}

func (s *Syncer) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Syncer) SetCapabilities(caps []string) {
	s.capabilities = caps
}

func (s *Syncer) agentRegistrationLoop(ctx context.Context) {
	defer s.wg.Done()

	// Stable agent ID: hostname-based, generated once per Syncer lifetime
	agentID := fmt.Sprintf("%s-%d", s.hostname, time.Now().UnixNano())

	// Retry registration until it succeeds
	for {
		select {
		case <-s.done:
			return
		default:
		}

		reg := AgentRegistration{
			AgentID:      agentID,
			Hostname:     s.hostname,
			Version:      s.version,
			Platform:     runtime.GOOS + "/" + runtime.GOARCH,
			Capabilities: s.capabilities,
		}

		if err := s.client.RegisterAgent(ctx, reg); err != nil {
			s.client.log("registration failed: %v (will retry)", err)
			time.Sleep(10 * time.Second)
			continue
		}

		s.mu.Lock()
		s.status = "registered"
		s.mu.Unlock()
		return
	}
}

func (s *Syncer) heartbeatLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.sendHeartbeat(ctx)
		}
	}
}

func (s *Syncer) sendHeartbeat(ctx context.Context) {
	hb := HeartbeatRequest{
		AgentID: s.client.AgentID(),
		Status:  "online",
	}

	if err := s.client.Heartbeat(ctx, hb); err != nil {
		s.client.log("heartbeat failed: %v", err)
	}
}

func (s *Syncer) queueReplayLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Try immediately on start
	s.replayQueue(ctx)

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.replayQueue(ctx)
		}
	}
}

func (s *Syncer) replayQueue(ctx context.Context) {
	q := s.client.Queue()
	if q == nil {
		return
	}

	items, err := q.DequeueAll()
	if err != nil {
		s.client.log("failed to dequeue: %v", err)
		return
	}

	if len(items) == 0 {
		return
	}

	s.mu.Lock()
	s.replaying = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.replaying = false
		s.mu.Unlock()
	}()

	s.client.log("replaying %d queued items", len(items))

	for _, item := range items {
		select {
		case <-s.done:
			return
		default:
		}

		if err := s.replayItem(ctx, item); err != nil {
			s.client.log("failed to replay %s (%s): %v", item.Type, item.ID, err)
			continue
		}

		if err := q.Remove(item.ID); err != nil {
			s.client.log("failed to remove replayed item %s: %v", item.ID, err)
		}
	}
}

func (s *Syncer) replayItem(ctx context.Context, item QueueItem) error {
	var target any

	switch item.Type {
	case "scan", "bulk_scans":
		target = &[]ScanPayload{}
	case "vulnerability", "bulk_vulnerabilities":
		target = &[]VulnerabilityPayload{}
	case "finding", "bulk_findings":
		target = &[]FindingPayload{}
	case "pipeline_event":
		target = &PipelineEventPayload{}
	case "heartbeat":
		target = &HeartbeatRequest{}
	case "agent_register":
		target = &AgentRegistration{}
	default:
		return nil
	}

	if err := json.Unmarshal(item.Payload, target); err != nil {
		// Skip corrupt items
		return nil
	}

	switch item.Type {
	case "scan":
		return s.client.UploadScan(ctx, *target.(*ScanPayload))
	case "bulk_scans":
		return s.client.UploadScansBulk(ctx, *target.(*[]ScanPayload))
	case "vulnerability":
		return s.client.UploadVulnerability(ctx, *target.(*VulnerabilityPayload))
	case "bulk_vulnerabilities":
		return s.client.UploadVulnerabilitiesBulk(ctx, *target.(*[]VulnerabilityPayload))
	case "finding":
		return s.client.UploadFinding(ctx, *target.(*FindingPayload))
	case "bulk_findings":
		return s.client.UploadFindingsBulk(ctx, *target.(*[]FindingPayload))
	case "pipeline_event":
		return s.client.UploadPipelineEvent(ctx, *target.(*PipelineEventPayload))
	case "heartbeat":
		return s.client.Heartbeat(ctx, *target.(*HeartbeatRequest))
	case "agent_register":
		return s.client.RegisterAgent(ctx, *target.(*AgentRegistration))
	}
	return nil
}

func (s *Syncer) SyncScanResults(ctx context.Context, tool string, raw json.RawMessage, workspaceID, projectID string) error {
	s.client.log("normalizing %s results", tool)

	results, err := NormalizeAll(tool, raw)
	if err != nil {
		return fmt.Errorf("normalization failed for %s: %w", tool, err)
	}

	if len(results) == 0 {
		s.client.log("no findings from %s", tool)
		return nil
	}

	s.client.log("normalized %d findings from %s", len(results), tool)

	scan := ScanPayload{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		ScanID:      fmt.Sprintf("%s-%d", tool, time.Now().Unix()),
		Type:        tool,
		Environment: "development",
		Status:      "completed",
		CompletedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.client.UploadScan(ctx, scan); err != nil {
		return err
	}

	var findings []FindingPayload
	var vulns []VulnerabilityPayload

	for _, r := range results {
		findings = append(findings, r.ToFindingPayload())
		vulns = append(vulns, r.ToVulnerabilityPayload(scan.ScanID))
	}

	if len(findings) > 0 {
		if err := s.client.UploadFindingsBulk(ctx, findings); err != nil {
			s.client.log("failed to upload %d findings: %v", len(findings), err)
		}
	}

	if len(vulns) > 0 {
		if err := s.client.UploadVulnerabilitiesBulk(ctx, vulns); err != nil {
			s.client.log("failed to upload %d vulnerabilities: %v", len(vulns), err)
		}
	}

	s.client.log("synced %d findings and %d vulns from %s", len(findings), len(vulns), tool)
	return nil
}

func (s *Syncer) SyncPipelineEvent(ctx context.Context, event PipelineEventPayload) error {
	if err := s.client.UploadPipelineEvent(ctx, event); err != nil {
		return err
	}
	return nil
}
