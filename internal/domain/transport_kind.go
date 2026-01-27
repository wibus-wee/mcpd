package domain

import "strings"

// NormalizeTransport canonicalizes transport kinds and defaults empty to stdio.
func NormalizeTransport(kind TransportKind) TransportKind {
	raw := strings.ToLower(strings.TrimSpace(string(kind)))
	switch raw {
	case "", string(TransportStdio):
		return TransportStdio
	case string(TransportStreamableHTTP), "streamable-http":
		return TransportStreamableHTTP
	default:
		return TransportKind(raw)
	}
}
