package registrybroker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetAgentFeedback returns the requested value.
func (c *RegistryBrokerClient) GetAgentFeedback(
	ctx context.Context,
	uaid string,
	options AgentFeedbackQuery,
) (JSONObject, error) {
	trimmed := strings.TrimSpace(uaid)
	if trimmed == "" {
		return nil, ensureNonEmpty(uaid, "uaid")
	}

	query := url.Values{}
	if options.IncludeRevoked {
		query.Set("includeRevoked", "true")
	}
	path := "/agents/" + percentPath(trimmed) + "/feedback"
	return c.requestJSON(ctx, http.MethodGet, pathWithQuery(path, query), nil, nil)
}

// ListAgentFeedbackIndex performs the requested operation.
func (c *RegistryBrokerClient) ListAgentFeedbackIndex(
	ctx context.Context,
	options AgentFeedbackIndexOptions,
) (JSONObject, error) {
	query := url.Values{}
	if options.Page != nil {
		query.Set("page", strconv.Itoa(*options.Page))
	}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	if len(options.Registries) > 0 {
		registries := make([]string, 0, len(options.Registries))
		for _, value := range options.Registries {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				registries = append(registries, trimmed)
			}
		}
		if len(registries) > 0 {
			query.Set("registry", strings.Join(registries, ","))
		}
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/agents/feedback", query),
		nil,
		nil,
	)
}

// ListAgentFeedbackEntriesIndex performs the requested operation.
func (c *RegistryBrokerClient) ListAgentFeedbackEntriesIndex(
	ctx context.Context,
	options AgentFeedbackIndexOptions,
) (JSONObject, error) {
	query := url.Values{}
	if options.Page != nil {
		query.Set("page", strconv.Itoa(*options.Page))
	}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	if len(options.Registries) > 0 {
		registries := make([]string, 0, len(options.Registries))
		for _, value := range options.Registries {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				registries = append(registries, trimmed)
			}
		}
		if len(registries) > 0 {
			query.Set("registry", strings.Join(registries, ","))
		}
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/agents/feedback/entries", query),
		nil,
		nil,
	)
}

// CheckAgentFeedbackEligibility performs the requested operation.
func (c *RegistryBrokerClient) CheckAgentFeedbackEligibility(
	ctx context.Context,
	uaid string,
	payload AgentFeedbackEligibilityRequest,
) (JSONObject, error) {
	trimmed := strings.TrimSpace(uaid)
	if trimmed == "" {
		return nil, ensureNonEmpty(uaid, "uaid")
	}
	path := "/agents/" + percentPath(trimmed) + "/feedback/eligibility"
	return c.requestJSON(
		ctx,
		http.MethodPost,
		path,
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// SubmitAgentFeedback submits the requested message payload.
func (c *RegistryBrokerClient) SubmitAgentFeedback(
	ctx context.Context,
	uaid string,
	payload AgentFeedbackSubmissionRequest,
) (JSONObject, error) {
	trimmed := strings.TrimSpace(uaid)
	if trimmed == "" {
		return nil, ensureNonEmpty(uaid, "uaid")
	}
	path := "/agents/" + percentPath(trimmed) + "/feedback"
	return c.requestJSON(
		ctx,
		http.MethodPost,
		path,
		payload,
		map[string]string{"content-type": "application/json"},
	)
}
