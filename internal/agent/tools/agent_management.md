# Agent Management Tools

Tools for managing sub-agent sessions.

## view_agent_graph

View the hierarchy of active agent sessions.

**Parameters:**
- None required

## create_agent

Create a new sub-agent session.

**Parameters:**
- `prompt` - The task for the agent to perform (required)
- `agent_type` - Type of agent to create (e.g., "task", "coder") (optional, defaults to "task")

## send_message_to_agent

Send a message to a running sub-agent session.

**Parameters:**
- `session_id` - The agent session ID to send message to (required)
- `message` - The message to send (required)

## agent_finish

Terminate a running sub-agent session.

**Parameters:**
- `session_id` - The agent session ID to terminate (required)