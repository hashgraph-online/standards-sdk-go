package registrybroker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *RegistryBrokerClient) Adapters(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/adapters", nil, nil)
}

func (c *RegistryBrokerClient) AdaptersDetailed(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/adapters/details", nil, nil)
}

func (c *RegistryBrokerClient) AdapterRegistryCategories(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/adapters/registry/categories", nil, nil)
}

func (c *RegistryBrokerClient) AdapterRegistryAdapters(
	ctx context.Context,
	filters AdapterRegistryFilters,
) (JSONObject, error) {
	query := url.Values{}
	addQueryString(query, "category", filters.Category)
	addQueryString(query, "entity", filters.Entity)
	addQueryString(query, "query", filters.Query)
	if filters.Limit != nil {
		query.Set("limit", strconv.Itoa(*filters.Limit))
	}
	if filters.Offset != nil {
		query.Set("offset", strconv.Itoa(*filters.Offset))
	}
	if len(filters.Keywords) > 0 {
		trimmed := make([]string, 0, len(filters.Keywords))
		for _, keyword := range filters.Keywords {
			value := strings.TrimSpace(keyword)
			if value != "" {
				trimmed = append(trimmed, value)
			}
		}
		if len(trimmed) > 0 {
			query.Set("keywords", strings.Join(trimmed, ","))
		}
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/adapters/registry/adapters", query),
		nil,
		nil,
	)
}

func (c *RegistryBrokerClient) CreateAdapterRegistryCategory(
	ctx context.Context,
	payload CreateAdapterRegistryCategoryRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/adapters/registry/categories",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) SubmitAdapterRegistryAdapter(
	ctx context.Context,
	payload SubmitAdapterRegistryAdapterRequest,
) (JSONObject, error) {
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/adapters/registry/adapters",
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) AdapterRegistrySubmissionStatus(
	ctx context.Context,
	submissionID string,
) (JSONObject, error) {
	if err := ensureNonEmpty(submissionID, "submissionID"); err != nil {
		return nil, err
	}
	path := "/adapters/registry/submissions/" + percentPath(submissionID)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}
