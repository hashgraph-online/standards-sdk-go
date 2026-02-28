package registrybroker

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRegistryBrokerIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" || os.Getenv("RUN_REGISTRY_BROKER_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 and RUN_REGISTRY_BROKER_INTEGRATION=1 to run live registry broker integration")
	}

	apiKey := os.Getenv("REGISTRY_BROKER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GODADDY_API_KEY")
	}
	if apiKey == "" {
		t.Skip("REGISTRY_BROKER_API_KEY (or GODADDY_API_KEY) is required for integration tests")
	}

	baseURL := os.Getenv("REGISTRY_BROKER_BASE_URL")
	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: baseURL,
		APIKey:  apiKey,
	})
	if err != nil {
		t.Fatalf("failed to create registry broker client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := client.Stats(ctx)
	if err != nil {
		t.Fatalf("stats request failed: %v", err)
	}
	if len(stats) == 0 {
		t.Fatalf("stats response should not be empty")
	}

	registries, err := client.Registries(ctx)
	if err != nil {
		t.Fatalf("registries request failed: %v", err)
	}
	if len(registries) == 0 {
		t.Fatalf("registries response should not be empty")
	}

	searchResult, err := client.Search(ctx, SearchParams{
		Q:     "agent",
		Limit: 1,
		Page:  1,
	})
	if err != nil {
		t.Fatalf("search request failed: %v", err)
	}
	if len(searchResult) == 0 {
		t.Fatalf("search response should not be empty")
	}
}
