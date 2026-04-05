package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSkillStatusParsesLifecycleSignals(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", request.Method)
		}
		if request.URL.Path != "/api/v1/skills/status" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		if request.URL.Query().Get("name") != "registry-broker" {
			t.Fatalf("unexpected name query %#v", request.URL.Query().Get("name"))
		}
		if request.URL.Query().Get("version") != "1.2.3" {
			t.Fatalf("unexpected version query %#v", request.URL.Query().Get("version"))
		}
		writer.Header().Set("content-type", "application/json")
		_, _ = writer.Write([]byte(`{
			"name": "registry-broker",
			"version": "1.2.3",
			"published": true,
			"verifiedDomain": true,
			"trustTier": "verified",
			"badgeMetric": "tier",
			"checks": {
				"repoCommitIntegrity": true,
				"manifestIntegrity": true,
				"domainProof": true
			},
			"nextSteps": [
				{
					"kind": "share_status",
					"priority": 1,
					"id": "share",
					"label": "Share status",
					"description": "Share the canonical page"
				}
			],
			"verificationSignals": {
				"publisherBound": true,
				"domainProof": true,
				"verifiedDomain": true,
				"previewValidated": true
			},
			"provenanceSignals": {
				"repoCommitIntegrity": true,
				"manifestIntegrity": true,
				"canonicalRelease": true,
				"previewAvailable": true,
				"previewAuthoritative": false
			},
			"statusUrl": "https://hol.org/registry/skills/registry-broker?version=1.2.3"
		}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	status, err := client.GetSkillStatus(
		context.Background(),
		SkillStatusRequest{Name: "registry-broker", Version: "1.2.3"},
	)
	if err != nil {
		t.Fatalf("get skill status failed: %v", err)
	}
	if status.TrustTier != SkillTrustTier("verified") {
		t.Fatalf("unexpected trust tier %#v", status.TrustTier)
	}
	if len(status.NextSteps) != 1 || status.NextSteps[0].Kind != "share_status" {
		t.Fatalf("unexpected next steps %#v", status.NextSteps)
	}
}

func TestSkillPreviewEndpointsUseCanonicalPaths(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch request.URL.Path {
		case "/api/v1/skills/status/by-repo":
			if request.URL.Query().Get("repo") != "https://github.com/hashgraph-online/registry-broker-skill" {
				t.Fatalf("unexpected repo query %#v", request.URL.Query().Get("repo"))
			}
			if request.URL.Query().Get("skillDir") != "." {
				t.Fatalf("unexpected skillDir query %#v", request.URL.Query().Get("skillDir"))
			}
			if request.URL.Query().Get("ref") != "refs/heads/main" {
				t.Fatalf("unexpected ref query %#v", request.URL.Query().Get("ref"))
			}
			_, _ = writer.Write([]byte(`{
				"name":"registry-broker",
				"version":"1.2.3",
				"published":false,
				"verifiedDomain":false,
				"trustTier":"validated",
				"badgeMetric":"tier",
				"checks":{"repoCommitIntegrity":true,"manifestIntegrity":true,"domainProof":false},
				"nextSteps":[],
				"verificationSignals":{"publisherBound":false,"domainProof":false,"verifiedDomain":false,"previewValidated":true},
				"provenanceSignals":{"repoCommitIntegrity":true,"manifestIntegrity":true,"canonicalRelease":false,"previewAvailable":true,"previewAuthoritative":false},
				"statusUrl":"https://hol.org/registry/skills/preview/preview-1"
			}`))
		case "/api/v1/skills/conversion-signals/by-repo":
			_, _ = writer.Write([]byte(`{
				"repoUrl":"https://github.com/hashgraph-online/registry-broker-skill",
				"skillDir":".",
				"trustTier":"validated",
				"actionInstalled":true,
				"previewUploaded":true,
				"previewId":"preview-1",
				"published":false,
				"verified":false,
				"publishReady":true,
				"publishBlockedByMissingAuth":false,
				"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
				"purchaseUrl":"https://hol.org/registry/skills/publish",
				"publishUrl":"https://hol.org/registry/skills/submit",
				"verificationUrl":"https://hol.org/registry/skills/registry-broker?tab=verification",
				"nextSteps":[]
			}`))
		case "/api/v1/skills/quote-preview":
			if request.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", request.Method)
			}
			_, _ = writer.Write([]byte(`{
				"estimatedCredits":{"min":589,"max":678},
				"estimatedHbar":{"min":4.11,"max":4.73},
				"pricingVersion":"2026-04-05",
				"assumptions":["2 files","3 KB total"],
				"purchaseUrl":"https://hol.org/registry/skills/publish",
				"publishUrl":"https://hol.org/registry/skills/submit",
				"verificationUrl":"https://hol.org/registry/skills/registry-broker?tab=verification"
			}`))
		case "/api/v1/skills/preview":
			_, _ = writer.Write([]byte(`{
				"found":true,
				"authoritative":false,
				"preview":{
					"id":"record-1",
					"previewId":"preview-1",
					"source":"github-oidc",
					"report":{
						"schema_version":"skill-preview.v1",
						"tool_version":"1.0.0",
						"preview_id":"preview-1",
						"repo_url":"https://github.com/hashgraph-online/registry-broker-skill",
						"repo_owner":"hashgraph-online",
						"repo_name":"registry-broker-skill",
						"default_branch":"main",
						"commit_sha":"abc123",
						"ref":"refs/pull/5/head",
						"event_name":"pull_request",
						"workflow_run_url":"https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
						"skill_dir":".",
						"name":"registry-broker",
						"version":"1.2.3",
						"validation_status":"passed",
						"findings":[],
						"package_summary":{"files":2},
						"suggested_next_steps":[],
						"generated_at":"2026-04-04T10:00:00.000Z"
					},
					"generatedAt":"2026-04-04T10:00:00.000Z",
					"expiresAt":"2026-04-11T10:00:00.000Z",
					"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
					"authoritative":false
				},
				"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
				"expiresAt":"2026-04-11T10:00:00.000Z"
			}`))
		case "/api/v1/skills/preview/by-repo":
			_, _ = writer.Write([]byte(`{
				"found":false,
				"authoritative":false
			}`))
		case "/api/v1/skills/preview/preview-1":
			_, _ = writer.Write([]byte(`{
				"found":true,
				"authoritative":false,
				"preview":{
					"id":"record-1",
					"previewId":"preview-1",
					"source":"github-oidc",
					"report":{
						"schema_version":"skill-preview.v1",
						"tool_version":"1.0.0",
						"preview_id":"preview-1",
						"repo_url":"https://github.com/hashgraph-online/registry-broker-skill",
						"repo_owner":"hashgraph-online",
						"repo_name":"registry-broker-skill",
						"default_branch":"main",
						"commit_sha":"abc123",
						"ref":"refs/pull/5/head",
						"event_name":"pull_request",
						"workflow_run_url":"https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
						"skill_dir":".",
						"name":"registry-broker",
						"version":"1.2.3",
						"validation_status":"passed",
						"findings":[],
						"package_summary":{"files":2},
						"suggested_next_steps":[],
						"generated_at":"2026-04-04T10:00:00.000Z"
					},
					"generatedAt":"2026-04-04T10:00:00.000Z",
					"expiresAt":"2026-04-11T10:00:00.000Z",
					"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
					"authoritative":false
				},
				"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
				"expiresAt":"2026-04-11T10:00:00.000Z"
			}`))
		case "/api/v1/skills/preview/github-oidc":
			if request.Header.Get("authorization") != "Bearer github-token" {
				t.Fatalf("unexpected authorization header %#v", request.Header.Get("authorization"))
			}
			_, _ = writer.Write([]byte(`{
				"id":"record-1",
				"previewId":"preview-1",
				"source":"github-oidc",
				"report":{
					"schema_version":"skill-preview.v1",
					"tool_version":"1.0.0",
					"preview_id":"preview-1",
					"repo_url":"https://github.com/hashgraph-online/registry-broker-skill",
					"repo_owner":"hashgraph-online",
					"repo_name":"registry-broker-skill",
					"default_branch":"main",
					"commit_sha":"abc123",
					"ref":"refs/pull/5/head",
					"event_name":"pull_request",
					"workflow_run_url":"https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
					"skill_dir":".",
					"name":"registry-broker",
					"version":"1.2.3",
					"validation_status":"passed",
					"findings":[],
					"package_summary":{"files":2},
					"suggested_next_steps":[],
					"generated_at":"2026-04-04T10:00:00.000Z"
				},
				"generatedAt":"2026-04-04T10:00:00.000Z",
				"expiresAt":"2026-04-11T10:00:00.000Z",
				"statusUrl":"https://hol.org/registry/skills/preview/preview-1",
				"authoritative":false
			}`))
		default:
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	statusByRepo, err := client.GetSkillStatusByRepo(context.Background(), SkillPreviewByRepoRequest{
		Repo:     "https://github.com/hashgraph-online/registry-broker-skill",
		SkillDir: ".",
		Ref:      "refs/heads/main",
	})
	if err != nil {
		t.Fatalf("get skill status by repo failed: %v", err)
	}
	if statusByRepo.TrustTier != "validated" {
		t.Fatalf("unexpected trust tier %#v", statusByRepo.TrustTier)
	}

	signals, err := client.GetSkillConversionSignalsByRepo(context.Background(), SkillPreviewByRepoRequest{
		Repo:     "https://github.com/hashgraph-online/registry-broker-skill",
		SkillDir: ".",
		Ref:      "refs/heads/main",
	})
	if err != nil {
		t.Fatalf("get skill conversion signals failed: %v", err)
	}
	if !signals.PublishReady {
		t.Fatalf("expected publish to be ready")
	}

	quote, err := client.QuoteSkillPublishPreview(context.Background(), SkillQuotePreviewRequest{
		FileCount:  2,
		TotalBytes: 3072,
		Name:       "registry-broker",
		Version:    "1.2.3",
		RepoURL:    "https://github.com/hashgraph-online/registry-broker-skill",
		SkillDir:   ".",
	})
	if err != nil {
		t.Fatalf("quote skill publish preview failed: %v", err)
	}
	if quote.EstimatedCredits.Min != 589 || quote.EstimatedCredits.Max != 678 {
		t.Fatalf("unexpected estimated credits %#v", quote.EstimatedCredits)
	}

	preview, err := client.GetSkillPreview(
		context.Background(),
		SkillPreviewLookupRequest{Name: "registry-broker", Version: "1.2.3"},
	)
	if err != nil {
		t.Fatalf("get skill preview failed: %v", err)
	}
	if preview.Preview == nil || preview.Preview.PreviewID != "preview-1" {
		t.Fatalf("unexpected preview %#v", preview.Preview)
	}

	previewByRepo, err := client.GetSkillPreviewByRepo(context.Background(), SkillPreviewByRepoRequest{
		Repo:     "https://github.com/hashgraph-online/registry-broker-skill",
		SkillDir: ".",
	})
	if err != nil {
		t.Fatalf("get skill preview by repo failed: %v", err)
	}
	if previewByRepo.Found {
		t.Fatalf("expected preview lookup by repo to be absent")
	}

	previewByID, err := client.GetSkillPreviewByID(context.Background(), "preview-1")
	if err != nil {
		t.Fatalf("get skill preview by id failed: %v", err)
	}
	if previewByID.Preview == nil || previewByID.Preview.ID != "record-1" {
		t.Fatalf("unexpected preview by id %#v", previewByID.Preview)
	}

	uploadedPreview, err := client.UploadSkillPreviewFromGitHubOIDC(context.Background(), "github-token", &SkillPreviewReport{
		SchemaVersion:      "skill-preview.v1",
		ToolVersion:        "1.0.0",
		PreviewID:          "preview-1",
		RepoURL:            "https://github.com/hashgraph-online/registry-broker-skill",
		RepoOwner:          "hashgraph-online",
		RepoName:           "registry-broker-skill",
		DefaultBranch:      "main",
		CommitSHA:          "abc123",
		Ref:                "refs/pull/5/head",
		EventName:          "pull_request",
		WorkflowRunURL:     "https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
		SkillDir:           ".",
		Name:               "registry-broker",
		Version:            "1.2.3",
		ValidationStatus:   "passed",
		Findings:           []any{},
		PackageSummary:     JSONObject{"files": 2},
		SuggestedNextSteps: []SkillPreviewSuggestedNextStep{},
		GeneratedAt:        "2026-04-04T10:00:00.000Z",
	})
	if err != nil {
		t.Fatalf("upload skill preview failed: %v", err)
	}
	if uploadedPreview.ID != "record-1" {
		t.Fatalf("unexpected uploaded preview %#v", uploadedPreview)
	}
}
