package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/SecDuckOps/duckops/internal/session"
)

const (
	ViewAgentGraphToolName     = "view_agent_graph"
	SendMessageToAgentToolName = "send_message_to_agent"
	AgentFinishToolName        = "agent_finish"
)

type ViewAgentGraphParams struct{}

type ViewAgentGraphResponseMetadata struct {
	Graph string `json:"graph"`
}

func NewViewAgentGraphTool(sessions session.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"view_agent_graph",
		"View the hierarchy of active agent sessions",
		func(ctx context.Context, params ViewAgentGraphParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			allSessions, err := sessions.List(ctx)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to list sessions: %w", err)
			}

			var agentSessions []session.Session
			for _, s := range allSessions {
				if sessions.IsAgentToolSession(s.ID) {
					agentSessions = append(agentSessions, s)
				}
			}

			if len(agentSessions) == 0 {
				return fantasy.NewTextResponse("No active agent sessions found."), nil
			}

			var b strings.Builder
			b.WriteString("# Agent Session Graph\n\n")
			b.WriteString("## Active Agent Sessions\n\n")

			for _, s := range agentSessions {
				b.WriteString(fmt.Sprintf("### %s\n\n", s.ID))
				b.WriteString(fmt.Sprintf("- **Title:** %s\n", s.Title))
				b.WriteString(fmt.Sprintf("- **Parent:** %s\n", s.ParentSessionID))
				b.WriteString(fmt.Sprintf("- **Messages:** %d\n", s.MessageCount))
				b.WriteString(fmt.Sprintf("- **Cost:** $%.4f\n", s.Cost))
				b.WriteString("\n")
			}

			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(b.String()),
				ViewAgentGraphResponseMetadata{Graph: b.String()},
			), nil
		})
}

type AgentFinishParams struct {
	SessionID string `json:"session_id" description:"The agent session ID to terminate"`
}

type AgentFinishResponseMetadata struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

func NewAgentFinishTool(sessions session.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"agent_finish",
		"Terminate a running sub-agent session",
		func(ctx context.Context, params AgentFinishParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.SessionID == "" {
				return fantasy.NewTextErrorResponse("session_id is required"), nil
			}

			sess, err := sessions.Get(ctx, params.SessionID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("session not found: %s", params.SessionID)), nil
			}

			if !sessions.IsAgentToolSession(sess.ID) {
				return fantasy.NewTextErrorResponse("not an agent session"), nil
			}

			err = sessions.Delete(ctx, params.SessionID)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to delete session: %w", err)
			}

			return fantasy.NewTextResponse(fmt.Sprintf("Agent session %s terminated successfully.", params.SessionID)), nil
		})
}

type SendMessageToAgentParams struct {
	SessionID string `json:"session_id" description:"The agent session ID to send message to"`
	Message   string `json:"message" description:"The message to send"`
}

type SendMessageToAgentResponseMetadata struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

func NewSendMessageToAgentTool(sessions session.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"send_message_to_agent",
		"Send a message to a running sub-agent session",
		func(ctx context.Context, params SendMessageToAgentParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.SessionID == "" {
				return fantasy.NewTextErrorResponse("session_id is required"), nil
			}
			if params.Message == "" {
				return fantasy.NewTextErrorResponse("message is required"), nil
			}

			sess, err := sessions.Get(ctx, params.SessionID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("session not found: %s", params.SessionID)), nil
			}

			if !sessions.IsAgentToolSession(sess.ID) {
				return fantasy.NewTextErrorResponse("not an agent session"), nil
			}

			return fantasy.NewTextResponse(fmt.Sprintf("Message sent to agent session %s: %s\n\nNote: This is a placeholder - agent message continuation not fully implemented.", params.SessionID, params.Message)), nil
		})
}