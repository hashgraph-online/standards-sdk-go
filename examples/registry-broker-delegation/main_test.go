package main

import "testing"

func TestGetEnvOrDefault(t *testing.T) {
	if value := getEnvOrDefault("REGISTRY_BROKER_DELEGATION_EXAMPLE_MISSING", "fallback"); value != "fallback" {
		t.Fatalf("expected fallback, got %s", value)
	}
}
