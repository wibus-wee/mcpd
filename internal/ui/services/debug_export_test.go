package services

import (
	"testing"
	"time"
)

func TestComputeStuck(t *testing.T) {
	now := time.Now().UTC()
	events := map[string][]diagnosticsEvent{
		"svc": {
			{
				Step:      "transport_connect",
				Phase:     "enter",
				Timestamp: now.Add(-2 * time.Minute).Format(time.RFC3339Nano),
			},
		},
		"fresh": {
			{
				Step:      "initialize_call",
				Phase:     "enter",
				Timestamp: now.Add(-5 * time.Second).Format(time.RFC3339Nano),
			},
		},
	}

	stuck := computeStuck(events, 30*time.Second)
	if _, ok := stuck["svc"]; !ok {
		t.Fatalf("expected svc to be stuck")
	}
	if _, ok := stuck["fresh"]; ok {
		t.Fatalf("expected fresh to be below threshold")
	}
}

func TestFormatAttributesForcedRedactions(t *testing.T) {
	attrs := map[string]string{
		"endpointSafe": "https://example.com/mcp",
	}
	sensitive := map[string]string{
		"cmd":      "./server --token=secret",
		"endpoint": "https://secret.example.com/mcp",
		"headers":  "{\"Authorization\":\"Bearer secret\"}",
		"env":      "{\"TOKEN\":\"secret\"}",
	}

	outSafe := formatAttributes(attrs, sensitive, "safe")
	if outSafe["cmd"] != "***" {
		t.Fatalf("expected cmd to be redacted in safe mode")
	}
	if outSafe["endpoint"] != "***" {
		t.Fatalf("expected endpoint to be redacted in safe mode")
	}
	if outSafe["headers"] != "***" {
		t.Fatalf("expected headers to be redacted in safe mode")
	}
	if _, ok := outSafe["env"]; ok {
		t.Fatalf("expected env to be omitted in safe mode")
	}
	if outSafe["endpointSafe"] != attrs["endpointSafe"] {
		t.Fatalf("expected endpointSafe to remain in safe mode")
	}

	outDeep := formatAttributes(attrs, sensitive, "deep")
	if outDeep["cmd"] != "***" {
		t.Fatalf("expected cmd to be redacted in deep mode")
	}
	if outDeep["endpoint"] != "***" {
		t.Fatalf("expected endpoint to be redacted in deep mode")
	}
	if outDeep["headers"] != "***" {
		t.Fatalf("expected headers to be redacted in deep mode")
	}
	if outDeep["env"] != sensitive["env"] {
		t.Fatalf("expected env to remain in deep mode")
	}
	if outDeep["endpointSafe"] != attrs["endpointSafe"] {
		t.Fatalf("expected endpointSafe to remain in deep mode")
	}
}
