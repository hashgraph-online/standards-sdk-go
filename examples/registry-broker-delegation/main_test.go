package main

import "testing"

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("returns fallback when env var is not set", func(t *testing.T) {
		if value := getEnvOrDefault("REGISTRY_BROKER_DELEGATION_EXAMPLE_MISSING", "fallback"); value != "fallback" {
			t.Fatalf("expected fallback, got %s", value)
		}
	})

	t.Run("returns env var value when set", func(t *testing.T) {
		t.Setenv("REGISTRY_BROKER_DELEGATION_EXAMPLE_SET", "env-value")

		if value := getEnvOrDefault("REGISTRY_BROKER_DELEGATION_EXAMPLE_SET", "fallback"); value != "env-value" {
			t.Fatalf("expected env-value, got %s", value)
		}
	})

	t.Run("returns trimmed env var value when set with whitespace", func(t *testing.T) {
		t.Setenv("REGISTRY_BROKER_DELEGATION_EXAMPLE_SPACED", "  env-value  ")

		if value := getEnvOrDefault("REGISTRY_BROKER_DELEGATION_EXAMPLE_SPACED", "fallback"); value != "env-value" {
			t.Fatalf("expected env-value, got %s", value)
		}
	})
}
