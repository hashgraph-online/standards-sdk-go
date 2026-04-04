package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const previewID = "preview-1"

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
			t.Fatalf("unexpected name query %s", request.URL.Query().Get("name"))
		}
		if request.URL.Query().Get("version") != "1.2.3" {
			t.Fatalf("unexpected version query %s", request.URL.Query().Get("version"))
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
					"id": "share-skill",
					"label": "Share status",
					"description": "Share the canonical page",
					"url": "https://hol.org/registry/skills/registry-broker"
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
			"publisher": {
				"cliPackageUrl": "https://www.npmjs.com/package/skill-publish",
				"cliCommand": "npx skill-publish",
				"actionMarketplaceUrl": "https://github.com/marketplace/actions/skill-publish",
				"repositoryUrl": "https://github.com/hashgraph-online/skill-publish",
				"quickstartCommands": [],
				"templatePresets": []
			},
			"preview": {
				"previewId": "preview-1",
				"repoUrl": "https://github.com/hashgraph-online/registry-broker-skill",
				"repoOwner": "hashgraph-online",
				"repoName": "registry-broker-skill",
				"commitSha": "abc123",
				"ref": "refs/pull/5/head",
				"eventName": "pull_request",
				"skillDir": ".",
				"generatedAt": "2026-04-04T10:00:00.000Z",
				"expiresAt": "2026-04-11T10:00:00.000Z",
				"statusUrl": "https://hol.org/registry/skills/preview/preview-1"
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
	if status.TrustTier != "verified" {
		t.Fatalf("unexpected trust tier %#v", status.TrustTier)
	}
	if !status.VerificationSignals.PreviewValidated {
		t.Fatalf("expected preview validated signal")
	}
	if status.Preview == nil || status.Preview.PreviewID != previewID {
		t.Fatalf("expected preview metadata %#v", status.Preview)
	}
	if len(status.NextSteps) != 1 || status.NextSteps[0].Kind != "share_status" {
		t.Fatalf("unexpected next steps %#v", status.NextSteps)
	}
}

func TestSkillPreviewEndpointsUseCanonicalPaths(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/skills/preview":
			if request.URL.Query().Get("name") != "registry-broker" {
				t.Fatalf("unexpected preview name query %s", request.URL.Query().Get("name"))
			}
			_, _ = writer.Write([]byte(`{
				"found": true,
				"authoritative": false,
				"preview": {
					"id": "record-1",
					"previewId": "preview-1",
					"source": "github-oidc",
					"report": {
						"schema_version": "skill-preview.v1",
						"tool_version": "1.0.0",
						"preview_id": "preview-1",
						"repo_url": "https://github.com/hashgraph-online/registry-broker-skill",
						"repo_owner": "hashgraph-online",
						"repo_name": "registry-broker-skill",
						"default_branch": "main",
						"commit_sha": "abc123",
						"ref": "refs/pull/5/head",
						"event_name": "pull_request",
						"workflow_run_url": "https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
						"skill_dir": ".",
						"name": "registry-broker",
						"version": "1.2.3",
						"validation_status": "passed",
						"findings": [],
						"package_summary": {"files": 2},
						"suggested_next_steps": [],
						"generated_at": "2026-04-04T10:00:00.000Z"
					},
					"generatedAt": "2026-04-04T10:00:00.000Z",
					"expiresAt": "2026-04-11T10:00:00.000Z",
					"statusUrl": "https://hol.org/registry/skills/preview/preview-1",
					"authoritative": false
				},
				"statusUrl": "https://hol.org/registry/skills/preview/preview-1",
				"expiresAt": "2026-04-11T10:00:00.000Z"
			}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/skills/preview/by-repo":
			if request.URL.Query().Get("repo") != "hashgraph-online/registry-broker-skill" {
				t.Fatalf("unexpected preview repo query %s", request.URL.Query().Get("repo"))
			}
			if request.URL.Query().Get("skillDir") != "." {
				t.Fatalf("unexpected preview skillDir query %s", request.URL.Query().Get("skillDir"))
			}
			_, _ = writer.Write([]byte(`{"found": false, "authoritative": false, "preview": null, "statusUrl": null, "expiresAt": null}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/skills/preview/preview-1":
			_, _ = writer.Write([]byte(`{"found": true, "authoritative": false, "preview": {"id":"record-1","previewId":"preview-1","source":"github-oidc","report":{"schema_version":"skill-preview.v1","tool_version":"1.0.0","preview_id":"preview-1","repo_url":"https://github.com/hashgraph-online/registry-broker-skill","repo_owner":"hashgraph-online","repo_name":"registry-broker-skill","default_branch":"main","commit_sha":"abc123","ref":"refs/pull/5/head","event_name":"pull_request","workflow_run_url":"https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1","skill_dir":".","name":"registry-broker","version":"1.2.3","validation_status":"passed","findings":[],"package_summary":{"files":2},"suggested_next_steps":[],"generated_at":"2026-04-04T10:00:00.000Z"},"generatedAt":"2026-04-04T10:00:00.000Z","expiresAt":"2026-04-11T10:00:00.000Z","statusUrl":"https://hol.org/registry/skills/preview/preview-1","authoritative":false},"statusUrl":"https://hol.org/registry/skills/preview/preview-1","expiresAt":"2026-04-11T10:00:00.000Z"}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/skills/preview/github-oidc":
			if !strings.HasPrefix(request.Header.Get("Authorization"), "Bearer github-token") {
				t.Fatalf("expected bearer token, got %q", request.Header.Get("Authorization"))
			}
			_, _ = writer.Write([]byte(`{
				"id": "record-1",
				"previewId": "preview-1",
				"source": "github-oidc",
				"report": {
					"schema_version": "skill-preview.v1",
					"tool_version": "1.0.0",
					"preview_id": "preview-1",
					"repo_url": "https://github.com/hashgraph-online/registry-broker-skill",
					"repo_owner": "hashgraph-online",
					"repo_name": "registry-broker-skill",
					"default_branch": "main",
					"commit_sha": "abc123",
					"ref": "refs/pull/5/head",
					"event_name": "pull_request",
					"workflow_run_url": "https://github.com/hashgraph-online/registry-broker-skill/actions/runs/1",
					"skill_dir": ".",
					"name": "registry-broker",
					"version": "1.2.3",
					"validation_status": "passed",
					"findings": [],
					"package_summary": {"files": 2},
					"suggested_next_steps": [],
					"generated_at": "2026-04-04T10:00:00.000Z"
				},
				"generatedAt": "2026-04-04T10:00:00.000Z",
				"expiresAt": "2026-04-11T10:00:00.000Z",
				"statusUrl": "https://hol.org/registry/skills/preview/preview-1",
				"authoritative": false
			}`))
		default:
			t.Fatalf("unexpected %s %s", request.Method, request.URL.Path)
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

	preview, err := client.GetSkillPreview(
		context.Background(),
		SkillPreviewLookupRequest{Name: "registry-broker", Version: "1.2.3"},
	)
	if err != nil {
		t.Fatalf("get skill preview failed: %v", err)
	}
	if !preview.Found || preview.Preview == nil || preview.Preview.PreviewID != previewID {
		t.Fatalf("unexpected preview %#v", preview)
	}

	byRepo, err := client.GetSkillPreviewByRepo(
		context.Background(),
		SkillPreviewByRepoRequest{
			Repo:     "hashgraph-online/registry-broker-skill",
			SkillDir: ".",
		},
	)
	if err != nil {
		t.Fatalf("get skill preview by repo failed: %v", err)
	}
	if byRepo.Found {
		t.Fatalf("expected preview by repo to be absent %#v", byRepo)
	}

	byID, err := client.GetSkillPreviewByID(context.Background(), previewID)
	if err != nil {
		t.Fatalf("get skill preview by id failed: %v", err)
	}
	if byID.Preview == nil || byID.Preview.ID != "record-1" {
		t.Fatalf("unexpected preview by id %#v", byID)
	}

	uploaded, err := client.UploadSkillPreviewFromGithubOIDC(
		context.Background(),
		"github-token",
		&SkillPreviewReport{
			SchemaVersion:      "skill-preview.v1",
			ToolVersion:        "1.0.0",
			PreviewID:          previewID,
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
		},
	)
	if err != nil {
		t.Fatalf("upload skill preview failed: %v", err)
	}
	if uploaded.Source != "github-oidc" || uploaded.PreviewID != previewID {
		t.Fatalf("unexpected uploaded preview %#v", uploaded)
	}
}
