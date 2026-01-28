package rpc

import "mcpd/internal/domain"

type ControlPlaneAPI interface {
	domain.InfoAPI
	domain.RegistryAPI
	domain.DiscoveryAPI
	domain.ObservabilityAPI
	domain.BootstrapAPI
	domain.AutomationAPI
	domain.SubAgentStatusAPI
	domain.TasksAPI
}
