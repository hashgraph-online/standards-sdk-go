package registrybroker

import (
	"context"
	"os"
	"testing"
)

func TestLiveSkillGrowthEndpoints(t *testing.T) {
	if os.Getenv("HASHNET_LIVE_TESTS") != "1" {
		t.Skip("set HASHNET_LIVE_TESTS=1 to run live registry broker checks")
	}

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: "https://hol.org/registry/api/v1",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	quote, err := client.QuoteSkillPublishPreview(context.Background(), &SkillQuotePreviewRequest{
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
	if quote.EstimatedCredits.Min <= 0 || quote.EstimatedCredits.Max <= 0 {
		t.Fatalf("unexpected quote response %#v", quote)
	}

	signals, err := client.GetSkillConversionSignalsByRepo(context.Background(), SkillPreviewByRepoRequest{
		Repo:     "https://github.com/hashgraph-online/registry-broker-skills",
		SkillDir: ".",
		Ref:      "refs/heads/main",
	})
	if err != nil {
		t.Fatalf("get skill conversion signals by repo failed: %v", err)
	}
	if signals.TrustTier == "" {
		t.Fatalf("expected trust tier in live response")
	}
}
