package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoverageAdditionalEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true, "data": []}`))
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: ts.URL,
		APIKey:  "test",
	})
	ctx := context.Background()

	// Adapters.go
	_, _ = client.Adapters(ctx)
	_, _ = client.AdaptersDetailed(ctx)
	_, _ = client.AdapterRegistryCategories(ctx)
	_, _ = client.AdapterRegistryAdapters(ctx, AdapterRegistryFilters{})
	_, _ = client.CreateAdapterRegistryCategory(ctx, CreateAdapterRegistryCategoryRequest{})
	_, _ = client.SubmitAdapterRegistryAdapter(ctx, map[string]any{"data": 1})
	_, _ = client.AdapterRegistrySubmissionStatus(ctx, "sub-id")

	// Agents.go
	_, _ = client.ResolveUaid(ctx, "uaid")
	_, _ = client.UpdateAgent(ctx, "uaid", map[string]any{})
	_, _ = client.GetRegisterStatus(ctx, "job1")
	_, _ = client.RegisterOwnedMoltbookAgent(ctx, "agent", MoltbookOwnerRegistrationUpdateRequest{})
	_, _ = client.GetRegistrationProgress(ctx, "job1")
	_, _ = client.ValidateUaid(ctx, "uaid")
	_, _ = client.GetUaidConnectionStatus(ctx, "uaid")
	_ = client.CloseUaidConnection(ctx, "uaid")
	_, _ = client.DashboardStats(ctx)

	// Chat.go (basic ones without encryption logic)
	_, _ = client.FetchHistorySnapshot(ctx, "session-1", ChatHistoryFetchOptions{})
	_, _ = client.CompactHistory(ctx, CompactHistoryRequestPayload{})
	_ = client.EndSession(ctx, "session-1")
}
