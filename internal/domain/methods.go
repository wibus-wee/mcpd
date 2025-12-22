package domain

func MethodAllowed(caps ServerCapabilities, method string) bool {
	switch method {
	case "ping":
		return true
	case "tools/list", "tools/call":
		return caps.Tools
	case "resources/list", "resources/read", "resources/subscribe", "resources/unsubscribe", "resources/templates/list":
		return caps.Resources
	case "prompts/list", "prompts/get":
		return caps.Prompts
	case "logging/setLevel", "notifications/message":
		return caps.Logging
	case "completion/complete":
		return caps.Completions
	default:
		return false
	}
}
