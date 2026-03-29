package registrybroker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDelegate(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/delegate" {
			t.Fatalf("expected /api/v1/delegate, got %s", r.URL.Path)
		}

		var request DelegationPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Task != "Review SDK PR feedback" {
			t.Fatalf("expected task to round-trip, got %q", request.Task)
		}
		if request.Filter == nil || len(request.Filter.Protocols) != 1 || request.Filter.Protocols[0] != "mcp" {
			t.Fatalf("expected protocols filter to round-trip, got %#v", request.Filter)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"task": "Review SDK PR feedback",
			"context": "Need a docs-focused sidecar pass.",
			"summary": "Delegate documentation follow-up.",
			"shouldDelegate": true,
			"localFirstReason": "Main agent owns the implementation work.",
			"recommendation": { "summary": "Delegate docs only", "mode": "parallel" },
			"opportunities": [
				{
					"id": "docs",
					"title": "Docs follow-up",
					"reason": "Bounded copy update",
					"role": "docs",
					"type": "sidecar",
					"suggestedMode": "parallel",
					"searchQueries": ["docs markdown docusaurus"],
					"extraOpportunityField": "preserved",
					"candidates": [
						{
							"uaid": "uaid-1",
							"label": "Docs Agent",
							"registry": "hcs-11",
							"score": 0.98,
							"trustScore": 0.91,
							"verified": true,
							"communicationSupported": true,
							"availability": "online",
							"explanation": "Strong docs match",
							"matchedRoles": ["docs"],
							"reasons": ["Strong docs match"],
							"suggestedMessage": "Update the docs tab set.",
							"extraCandidateField": "preserved",
							"agent": {
								"name": "Docs Agent",
								"verified": true,
								"extraAgentField": "preserved"
							}
						}
					]
				}
			],
			"extraRootField": "preserved"
		}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.Delegate(context.Background(), DelegationPlanRequest{
		Task:    "Review SDK PR feedback",
		Context: "Need a docs-focused sidecar pass.",
		Limit:   2,
		Filter: &DelegationPlanFilter{
			Protocols: []string{"mcp"},
			Type:      "mcp_server",
		},
		Workspace: JSONObject{"repo": "hashgraph-online/standards-sdk"},
	})
	if err != nil {
		t.Fatalf("delegate: %v", err)
	}

	if !response.ShouldDelegate {
		t.Fatal("expected delegation recommendation")
	}
	if response.Context != "Need a docs-focused sidecar pass." {
		t.Fatalf("expected context to parse, got %q", response.Context)
	}
	if len(response.Opportunities) != 1 {
		t.Fatalf("expected one opportunity, got %d", len(response.Opportunities))
	}
	if response.Extras["extraRootField"] != "preserved" {
		t.Fatalf("expected additive root field to survive, got %#v", response.Extras)
	}
	opportunity := response.Opportunities[0]
	if opportunity.Extras["extraOpportunityField"] != "preserved" {
		t.Fatalf("expected additive opportunity field to survive, got %#v", opportunity.Extras)
	}
	candidate := opportunity.Candidates[0]
	if candidate.Registry != "hcs-11" {
		t.Fatalf("expected registry to parse, got %q", candidate.Registry)
	}
	if candidate.TrustScore != 0.91 {
		t.Fatalf("expected trust score 0.91, got %f", candidate.TrustScore)
	}
	if candidate.Verified == nil || !*candidate.Verified {
		t.Fatal("expected verified candidate")
	}
	if candidate.Extras["extraCandidateField"] != "preserved" {
		t.Fatalf("expected additive candidate field to survive, got %#v", candidate.Extras)
	}
	if candidate.Agent["extraAgentField"] != "preserved" {
		t.Fatalf("expected additive agent field to survive, got %#v", candidate.Agent)
	}
}
