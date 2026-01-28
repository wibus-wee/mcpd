package app

import "mcpd/internal/domain"

// ControlPlaneAPI describes the surface exposed to external coordinators (e.g. UI).
type ControlPlaneAPI interface {
	domain.InfoAPI
	domain.RegistryAPI
	domain.DiscoveryAPI
	domain.ObservabilityAPI
	domain.BootstrapAPI
	domain.SubAgentStatusAPI
	domain.StoreAPI
}
