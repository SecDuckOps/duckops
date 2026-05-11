package agent

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"charm.land/fantasy"

	"github.com/SecDuckOps/duckops/internal/agent/prompt"
	"github.com/SecDuckOps/duckops/internal/agent/tools"
	"github.com/SecDuckOps/duckops/internal/config"
)

//go:embed templates/agent_tool.md
var agentToolDescription []byte

type AgentParams struct {
	Prompt string `json:"prompt" description:"The task for the agent to perform"`
}

type CreateAgentParams struct {
	Prompt    string `json:"prompt" description:"The task for the agent to perform"`
	AgentType string `json:"agent_type" description:"Type of agent to create (task, coder)"`
}

const (
	AgentToolName = "agent"
)

func (c *coordinator) agentTool(ctx context.Context) (fantasy.AgentTool, error) {
	agentCfg, ok := c.cfg.Config().Agents[config.AgentTask]
	if !ok {
		return nil, errors.New("task agent not configured")
	}
	prompt, err := taskPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
	if err != nil {
		return nil, err
	}

	agent, err := c.buildAgent(ctx, prompt, agentCfg, true)
	if err != nil {
		return nil, err
	}
	return fantasy.NewParallelAgentTool(
		AgentToolName,
		tools.FirstLineDescription(agentToolDescription),
		func(ctx context.Context, params AgentParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Prompt == "" {
				return fantasy.NewTextErrorResponse("prompt is required"), nil
			}

			sessionID := tools.GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, errors.New("session id missing from context")
			}

			agentMessageID := tools.GetMessageFromContext(ctx)
			if agentMessageID == "" {
				return fantasy.ToolResponse{}, errors.New("agent message id missing from context")
			}

			return c.runSubAgent(ctx, subAgentParams{
				Agent:          agent,
				SessionID:      sessionID,
				AgentMessageID: agentMessageID,
				ToolCallID:     call.ID,
				Prompt:         params.Prompt,
				SessionTitle:   "New Agent Session",
			})
		}), nil
}

func (c *coordinator) createAgentTool(ctx context.Context) (fantasy.AgentTool, error) {
	return fantasy.NewAgentTool(
		"create_agent",
		"Create and run a new sub-agent session",
		func(ctx context.Context, params CreateAgentParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Prompt == "" {
				return fantasy.NewTextErrorResponse("prompt is required"), nil
			}

			sessionID := tools.GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, errors.New("session id missing from context")
			}

			agentMessageID := tools.GetMessageFromContext(ctx)
			if agentMessageID == "" {
				return fantasy.ToolResponse{}, errors.New("agent message id missing from context")
			}

			agentType := params.AgentType
			if agentType == "" {
				agentType = "task"
			}

			agentCfg, ok := c.cfg.Config().Agents[agentType]
			if !ok {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("agent type %q not configured", agentType)), nil
			}

			agentPrompt, err := taskPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to create prompt: %v", err)), nil
			}

			agent, err := c.buildAgent(ctx, agentPrompt, agentCfg, true)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to build agent: %v", err)), nil
			}

			title := params.Prompt
			if len(title) > 50 {
				title = title[:50]
			}
			return c.runSubAgent(ctx, subAgentParams{
				Agent:          agent,
				SessionID:      sessionID,
				AgentMessageID: agentMessageID,
				ToolCallID:     call.ID,
				Prompt:         params.Prompt,
				SessionTitle:   "Agent: " + title,
			})
		}), nil
}
