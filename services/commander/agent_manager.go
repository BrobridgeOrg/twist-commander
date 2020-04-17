package commander

import app "twist-commander/app/interface"

type AgentManager struct {
	app app.AppImpl
}

func CreateAgentManager(a app.AppImpl) *AgentManager {
	return &AgentManager{
		app: a,
	}
}

func (am *AgentManager) CreateAgent(transactionID string) *Agent {

	agent := CreateAgent(am.app, transactionID)

	return agent
}
