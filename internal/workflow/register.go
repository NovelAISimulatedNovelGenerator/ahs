package workflow

import (
	simple_example "ahs/internal/workflow/eino_imp/simple_example"
)

func (m *Manager) RegisterWorkflows() {
	// Register workflow under here:
	m.Register("simple_example", &simple_example.SimpleProcessor{})
}
