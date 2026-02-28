package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoverageAllEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true, "data": []}`))
	}))
	defer ts.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: ts.URL,
		APIKey:  "test",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	ctx := context.Background()

	// feedback
	_, _ = client.GetAgentFeedback(ctx, "uaid", AgentFeedbackQuery{})
	_, _ = client.CheckAgentFeedbackEligibility(ctx, "uaid", AgentFeedbackEligibilityRequest{})
	_, _ = client.SubmitAgentFeedback(ctx, "uaid", AgentFeedbackSubmissionRequest{})

	// search
	_, _ = client.SearchErc8004ByAgentID(ctx, 1, "uaid", nil, nil, "", "")
	_, _ = client.Stats(ctx)
	_, _ = client.Registries(ctx)
	_, _ = client.GetAdditionalRegistries(ctx)
	_, _ = client.PopularSearches(ctx)
	_, _ = client.ListProtocols(ctx)
	_, _ = client.DetectProtocol(ctx, JSONObject{})
	_, _ = client.RegistrySearchByNamespace(ctx, "ans", "query")
	_, _ = client.VectorSearch(ctx, VectorSearchRequest{})
	_, _ = client.SearchStatus(ctx)
	_, _ = client.WebsocketStats(ctx)
	_, _ = client.MetricsSummary(ctx)
	// `client.Facets` takes (ctx, registry string) or similar based on the error "cannot use SearchParams{} as string" 
	_, _ = client.Facets(ctx, "ans")

	// skills
	_, _ = client.GetSkillsCatalog(ctx, SkillCatalogOptions{})
	_, _ = client.ListSkills(ctx, ListSkillsOptions{})
	_, _ = client.ListSkillVersions(ctx, "skill-id")
	_, _ = client.ListMySkills(ctx, ListMySkillsOptions{})
	_, _ = client.GetMySkillsList(ctx, MySkillsListOptions{})
	_, _ = client.QuoteSkillPublish(ctx, SkillRegistryQuoteRequest{})
	_, _ = client.PublishSkill(ctx, SkillRegistryPublishRequest{})
	_, _ = client.GetSkillPublishJob(ctx, "job1", SkillPublishJobOptions{})
	_, _ = client.GetSkillOwnership(ctx, "skill-id", "owner-id")
	_, _ = client.GetRecommendedSkillVersion(ctx, "skill-id")
	_, _ = client.SetRecommendedSkillVersion(ctx, SkillRecommendedVersionSetRequest{})
	_, _ = client.GetSkillDeprecations(ctx, "skill-id")
	_, _ = client.SetSkillDeprecation(ctx, SkillDeprecationSetRequest{})
	_, _ = client.GetSkillVoteStatus(ctx, "skill-id")
	_, _ = client.SetSkillVote(ctx, SkillRegistryVoteRequest{})
	_, _ = client.RequestSkillVerification(ctx, SkillVerificationRequestCreateRequest{})
	_, _ = client.GetSkillVerificationStatus(ctx, "skill-id")

	// verification
	_, _ = client.GetVerificationStatus(ctx, "target")
	_, _ = client.CreateVerificationChallenge(ctx, "target")
	_, _ = client.GetVerificationChallenge(ctx, "chal")
	_, _ = client.VerifyVerificationChallenge(ctx, VerifyVerificationChallengeRequest{})
	_, _ = client.GetVerificationOwnership(ctx, "target")
	_, _ = client.VerifySenderOwnership(ctx, "uaid")

	// credits
	_, _ = client.PurchaseCreditsWithX402(ctx, PurchaseCreditsWithX402Params{})
	_, _ = client.BuyCreditsWithX402(ctx, BuyCreditsWithX402Params{})
}
