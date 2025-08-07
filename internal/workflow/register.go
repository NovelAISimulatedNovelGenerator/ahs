package workflow

import (
	agent "ahs/internal/workflow/eino_imp/agent"
	simple_example "ahs/internal/workflow/eino_imp/simple_example"
)

func (m *Manager) RegisterWorkflows() {
	// Register workflow under here:
	m.Register("agent", &agent.AgentProcessor{})
	m.Register("simple_example", &simple_example.SimpleProcessor{})
}
