package loader

import (
	"github.com/spf13/viper"

	"mcpv/internal/domain"
)

func newRuntimeViper() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	setRuntimeDefaults(v)
	return v
}

func setRuntimeDefaults(v *viper.Viper) {
	v.SetDefault("routeTimeoutSeconds", domain.DefaultRouteTimeoutSeconds)
	v.SetDefault("pingIntervalSeconds", domain.DefaultPingIntervalSeconds)
	v.SetDefault("toolRefreshSeconds", domain.DefaultToolRefreshSeconds)
	v.SetDefault("toolRefreshConcurrency", domain.DefaultToolRefreshConcurrency)
	v.SetDefault("clientCheckSeconds", domain.DefaultClientCheckSeconds)
	v.SetDefault("clientInactiveSeconds", domain.DefaultClientInactiveSeconds)
	v.SetDefault("serverInitRetryBaseSeconds", domain.DefaultServerInitRetryBaseSeconds)
	v.SetDefault("serverInitRetryMaxSeconds", domain.DefaultServerInitRetryMaxSeconds)
	v.SetDefault("serverInitMaxRetries", domain.DefaultServerInitMaxRetries)
	v.SetDefault("bootstrapMode", domain.DefaultBootstrapMode)
	v.SetDefault("bootstrapConcurrency", domain.DefaultBootstrapConcurrency)
	v.SetDefault("bootstrapTimeoutSeconds", domain.DefaultBootstrapTimeoutSeconds)
	v.SetDefault("defaultActivationMode", domain.DefaultActivationMode)
	v.SetDefault("exposeTools", domain.DefaultExposeTools)
	v.SetDefault("toolNamespaceStrategy", domain.DefaultToolNamespaceStrategy)
	v.SetDefault("observability.listenAddress", domain.DefaultObservabilityListenAddress)
	v.SetDefault("rpc.listenAddress", domain.DefaultRPCListenAddress)
	v.SetDefault("rpc.maxRecvMsgSize", domain.DefaultRPCMaxRecvMsgSize)
	v.SetDefault("rpc.maxSendMsgSize", domain.DefaultRPCMaxSendMsgSize)
	v.SetDefault("rpc.keepaliveTimeSeconds", domain.DefaultRPCKeepaliveTimeSeconds)
	v.SetDefault("rpc.keepaliveTimeoutSeconds", domain.DefaultRPCKeepaliveTimeoutSeconds)
	v.SetDefault("rpc.socketMode", domain.DefaultRPCSocketMode)
}
