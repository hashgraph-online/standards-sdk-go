package registrybroker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCovSearchHelpers(t *testing.T) {
	result := buildVectorFallbackSearchParams(VectorSearchRequest{
		Query:  "test",
		Limit:  10,
		Offset: 20,
		Filter: &VectorSearchFilter{
			Registry:     "hcs-1",
			Protocols:    []string{"p1"},
			Adapter:      []string{"a1"},
			Capabilities: []string{"c1"},
			Type:         "ai_agent",
		},
	})
	if result.Q != "test" { t.Fatal("expected test") }
	if result.Page != 3 { t.Fatalf("expected page 3 got %d", result.Page) }
	if result.Registry != "hcs-1" { t.Fatal("expected hcs-1") }

	result2 := convertSearchResultToVectorResponse(JSONObject{
		"hits":  []any{JSONObject{"name": "a1"}},
		"total": float64(10),
		"limit": float64(5),
		"page":  float64(2),
	})
	if result2["total"] != float64(10) { t.Fatal("expected 10") }
}

func TestCovStringifyValue(t *testing.T) {
	if stringifyValue("hello") != "hello" { t.Fatal("expected hello") }
	if stringifyValue(42) != "42" { t.Fatal("expected 42") }
	if stringifyValue(int64(100)) != "100" { t.Fatal("expected 100") }
	if stringifyValue(3.14) != "3.14" { t.Fatal("expected 3.14") }
	if stringifyValue(true) != "true" { t.Fatal("expected true") }
	if stringifyValue(nil) != "<nil>" { t.Fatal("expected nil str") }
}

func TestCovGetNumberField(t *testing.T) {
	m := JSONObject{"f": 2.5, "i": float64(10), "s": "str"}
	v, ok := getNumberField(m, "f")
	if !ok || v != 2.5 { t.Fatal("expected 2.5") }

	_, ok2 := getNumberField(m, "missing")
	if ok2 { t.Fatal("expected false") }

	_, ok3 := getNumberField(m, "s")
	if ok3 { t.Fatal("expected false") }
}

func TestCovFeedbackWithFilters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: ts.URL})
	ctx := context.Background()

	page := 1
	limit := 10
	_, _ = client.ListAgentFeedbackIndex(ctx, AgentFeedbackIndexOptions{
		Page:       &page,
		Limit:      &limit,
		Registries: []string{"hcs-1", ""},
	})
	_, _ = client.ListAgentFeedbackEntriesIndex(ctx, AgentFeedbackIndexOptions{
		Page:       &page,
		Limit:      &limit,
		Registries: []string{"hcs-1"},
	})
}

func TestCovCompactHistory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: ts.URL})
	ctx := context.Background()

	_, err := client.CompactHistory(ctx, CompactHistoryRequestPayload{SessionID: ""})
	if err == nil { t.Fatal("expected error") }

	preserve := 5
	_, _ = client.CompactHistory(ctx, CompactHistoryRequestPayload{
		SessionID:       "s1",
		PreserveEntries: &preserve,
	})
}

func TestCovEncryptionUnavailableError(t *testing.T) {
	e := &EncryptionUnavailableError{SessionID: "s1"}
	if e.Error() == "" { t.Fatal("expected error string") }
}

func TestCovBuildSearchQuery(t *testing.T) {
	verified := true
	online := true
	minTrust := 0.5
	q := buildSearchQuery(SearchParams{
		Q:            "test",
		Page:         1,
		Limit:        10,
		Registry:     "hcs-1",
		Registries:   []string{"hcs-1", "hcs-2"},
		Capabilities: []string{"c1"},
		Protocols:    []string{"p1"},
		Adapters:     []string{"a1"},
		MinTrust:     &minTrust,
		Metadata: map[string][]any{
			"key":  {"value"},
			"":     {"ignored"},
			"key2": {nil},
		},
		Type:      "ai_agent",
		Verified:  &verified,
		Online:    &online,
		SortBy:    "name",
		SortOrder: "asc",
	})
	if q == "" { t.Fatal("expected non-empty query") }

	q2 := buildSearchQuery(SearchParams{})
	if q2 != "" { t.Fatal("expected empty") }
}

func TestCovSearchEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: ts.URL})
	ctx := context.Background()

	_, _ = client.Stats(ctx)
	_, _ = client.Registries(ctx)
	_, _ = client.GetAdditionalRegistries(ctx)
	_, _ = client.PopularSearches(ctx)
	_, _ = client.ListProtocols(ctx)
	_, _ = client.DetectProtocol(ctx, JSONObject{"test": true})
	_, _ = client.SearchStatus(ctx)
	_, _ = client.WebsocketStats(ctx)
	_, _ = client.MetricsSummary(ctx)
	_, _ = client.Facets(ctx, "adapter1")
	_, _ = client.RegistrySearchByNamespace(ctx, "hcs-1", "query")
	_, _ = client.RegistrySearchByNamespace(ctx, "", "query")
	_, _ = client.VectorSearch(ctx, VectorSearchRequest{Query: "test"})
	_, _ = client.SearchErc8004ByAgentID(ctx, 1, "agent1", nil, nil, "", "")
	_, _ = client.SearchErc8004ByAgentID(ctx, 0, "agent1", nil, nil, "", "")
	_, _ = client.SearchErc8004ByAgentID(ctx, 1, "", nil, nil, "", "")
}
