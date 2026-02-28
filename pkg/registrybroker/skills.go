package registrybroker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *RegistryBrokerClient) SkillsConfig(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/skills/config", nil, nil)
}

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

func (c *RegistryBrokerClient) ListSkills(
	ctx context.Context,
	options ListSkillsOptions,
) (JSONObject, error) {
	query := url.Values{}
	addQueryString(query, "name", options.Name)
	addQueryString(query, "version", options.Version)
	addQueryString(query, "cursor", options.Cursor)
	addQueryString(query, "accountId", options.AccountID)
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

func (c *RegistryBrokerClient) GetSkillVerificationStatus(
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
		pathWithQuery("/skills/verification/status", query),
		nil,
		nil,
	)
}
