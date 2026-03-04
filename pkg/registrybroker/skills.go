package registrybroker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SkillsConfig performs the requested operation.
func (c *RegistryBrokerClient) SkillsConfig(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/skills/config", nil, nil)
}

// GetSkillsCatalog returns the requested value.
func (c *RegistryBrokerClient) GetSkillsCatalog(
	ctx context.Context,
	options SkillCatalogOptions,
) (JSONObject, error) {
	query := url.Values{}
	addQueryString(query, "q", options.Q)
	addQueryString(query, "category", options.Category)
	addQueryString(query, "channel", options.Channel)
	addQueryString(query, "sortBy", options.SortBy)
	addQueryString(query, "cursor", options.Cursor)
	if options.Featured != nil {
		query.Set("featured", strconv.FormatBool(*options.Featured))
	}
	if options.Verified != nil {
		query.Set("verified", strconv.FormatBool(*options.Verified))
	}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	for _, tag := range options.Tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			query.Add("tag", trimmed)
		}
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/catalog", query),
		nil,
		nil,
	)
}

// ListSkills performs the requested operation.
func (c *RegistryBrokerClient) ListSkills(
	ctx context.Context,
	options ListSkillsOptions,
) (JSONObject, error) {
	query := url.Values{}
	addQueryString(query, "name", options.Name)
	addQueryString(query, "version", options.Version)
	addQueryString(query, "q", options.Q)
	addQueryString(query, "tag", options.Tag)
	addQueryString(query, "category", options.Category)
	addQueryString(query, "view", options.View)
	addQueryString(query, "cursor", options.Cursor)
	addQueryString(query, "accountId", options.AccountID)
	addQueryBool(query, "featured", options.Featured)
	addQueryBool(query, "verified", options.Verified)
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	if options.IncludeFiles != nil {
		query.Set("includeFiles", strconv.FormatBool(*options.IncludeFiles))
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills", query),
		nil,
		nil,
	)
}

// GetSkillSecurityBreakdown returns scanner summary and findings for a skill release.
func (c *RegistryBrokerClient) GetSkillSecurityBreakdown(
	ctx context.Context,
	options SkillSecurityBreakdownOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(options.JobID, "jobID"); err != nil {
		return nil, err
	}

	path := "/skills/" + percentPath(options.JobID) + "/security-breakdown"

	return c.requestJSON(
		ctx,
		http.MethodGet,
		path,
		nil,
		nil,
	)
}

// ListSkillVersions performs the requested operation.
func (c *RegistryBrokerClient) ListSkillVersions(
	ctx context.Context,
	name string,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/versions", query),
		nil,
		nil,
	)
}

// ListMySkills performs the requested operation.
func (c *RegistryBrokerClient) ListMySkills(
	ctx context.Context,
	options ListMySkillsOptions,
) (JSONObject, error) {
	query := url.Values{}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/mine", query),
		nil,
		nil,
	)
}

// GetMySkillsList returns the requested value.
func (c *RegistryBrokerClient) GetMySkillsList(
	ctx context.Context,
	options MySkillsListOptions,
) (JSONObject, error) {
	query := url.Values{}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	addQueryString(query, "cursor", options.Cursor)
	addQueryString(query, "accountId", options.AccountID)
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/my-list", query),
		nil,
		nil,
	)
}

// QuoteSkillPublish performs the requested operation.
func (c *RegistryBrokerClient) QuoteSkillPublish(
	ctx context.Context,
	payload SkillRegistryQuoteRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/quote",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// PublishSkill publishes the requested message payload.
func (c *RegistryBrokerClient) PublishSkill(
	ctx context.Context,
	payload SkillRegistryPublishRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/publish",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// GetSkillPublishJob returns the requested value.
func (c *RegistryBrokerClient) GetSkillPublishJob(
	ctx context.Context,
	jobID string,
	options SkillPublishJobOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(jobID, "jobID"); err != nil {
		return nil, err
	}
	query := url.Values{}
	addQueryString(query, "accountId", options.AccountID)
	path := "/skills/jobs/" + percentPath(jobID)
	return c.requestJSON(ctx, http.MethodGet, pathWithQuery(path, query), nil, nil)
}

// GetSkillOwnership returns the requested value.
func (c *RegistryBrokerClient) GetSkillOwnership(
	ctx context.Context,
	name string,
	accountID string,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	addQueryString(query, "accountId", accountID)
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/ownership", query),
		nil,
		nil,
	)
}

// GetRecommendedSkillVersion returns the requested value.
func (c *RegistryBrokerClient) GetRecommendedSkillVersion(
	ctx context.Context,
	name string,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/recommended", query),
		nil,
		nil,
	)
}

// SetRecommendedSkillVersion sets the requested value.
func (c *RegistryBrokerClient) SetRecommendedSkillVersion(
	ctx context.Context,
	payload SkillRecommendedVersionSetRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/recommended",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// GetSkillDeprecations returns the requested value.
func (c *RegistryBrokerClient) GetSkillDeprecations(
	ctx context.Context,
	name string,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/deprecations", query),
		nil,
		nil,
	)
}

// SetSkillDeprecation sets the requested value.
func (c *RegistryBrokerClient) SetSkillDeprecation(
	ctx context.Context,
	payload SkillDeprecationSetRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/deprecate",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// GetSkillBadge returns the requested value.
func (c *RegistryBrokerClient) GetSkillBadge(
	ctx context.Context,
	options SkillBadgeOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(options.Name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(options.Name))
	addQueryString(query, "metric", options.Metric)
	addQueryString(query, "label", options.Label)
	addQueryString(query, "style", options.Style)
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/badge", query),
		nil,
		nil,
	)
}

// ListSkillTags returns the requested value.
func (c *RegistryBrokerClient) ListSkillTags(
	ctx context.Context,
) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/skills/tags", nil, nil)
}

// ListSkillCategories returns the requested value.
func (c *RegistryBrokerClient) ListSkillCategories(
	ctx context.Context,
) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/skills/categories", nil, nil)
}

// ResolveSkillMarkdown returns the SKILL.md payload for a skill reference.
func (c *RegistryBrokerClient) ResolveSkillMarkdown(
	ctx context.Context,
	skillRef string,
) (string, error) {
	if err := ensureNonEmpty(skillRef, "skillRef"); err != nil {
		return "", err
	}
	path := "/skills/" + percentPath(skillRef) + "/SKILL.md"
	body, _, err := c.request(ctx, http.MethodGet, path, nil, map[string]string{
		"accept": "text/markdown, text/plain;q=0.9, */*;q=0.8",
	})
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ResolveSkillManifest returns the SKILL.json manifest for a skill reference.
func (c *RegistryBrokerClient) ResolveSkillManifest(
	ctx context.Context,
	skillRef string,
) (JSONObject, error) {
	if err := ensureNonEmpty(skillRef, "skillRef"); err != nil {
		return nil, err
	}
	path := "/skills/" + percentPath(skillRef) + "/manifest"
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// GetSkillVoteStatus returns the requested value.
func (c *RegistryBrokerClient) GetSkillVoteStatus(
	ctx context.Context,
	name string,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/vote", query),
		nil,
		nil,
	)
}

// SetSkillVote sets the requested value.
func (c *RegistryBrokerClient) SetSkillVote(
	ctx context.Context,
	payload SkillRegistryVoteRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/vote",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// RequestSkillVerification performs the requested operation.
func (c *RegistryBrokerClient) RequestSkillVerification(
	ctx context.Context,
	payload SkillVerificationRequestCreateRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/verification/request",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// GetSkillVerificationStatus returns the requested value.
func (c *RegistryBrokerClient) GetSkillVerificationStatus(
	ctx context.Context,
	name string,
) (JSONObject, error) {
	return c.GetSkillVerificationStatusWithOptions(
		ctx,
		name,
		SkillVerificationStatusOptions{},
	)
}

// GetSkillVerificationStatusWithOptions returns the requested value.
func (c *RegistryBrokerClient) GetSkillVerificationStatusWithOptions(
	ctx context.Context,
	name string,
	options SkillVerificationStatusOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(name, "name"); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("name", strings.TrimSpace(name))
	addQueryString(query, "version", options.Version)
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/skills/verification/status", query),
		nil,
		nil,
	)
}

// CreateSkillDomainProofChallenge creates a DNS TXT verification challenge.
func (c *RegistryBrokerClient) CreateSkillDomainProofChallenge(
	ctx context.Context,
	payload SkillVerificationDomainProofChallengeRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/verification/domain/challenge",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// VerifySkillDomainProof validates the DNS TXT challenge response.
func (c *RegistryBrokerClient) VerifySkillDomainProof(
	ctx context.Context,
	payload SkillVerificationDomainProofVerifyRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/skills/verification/domain/verify",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}
