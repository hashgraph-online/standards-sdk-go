package registrybroker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Search performs the requested operation.
func (c *RegistryBrokerClient) Search(ctx context.Context, params SearchParams) (JSONObject, error) {
	path := "/search" + buildSearchQuery(params)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// SearchErc8004ByAgentID performs the requested operation.
func (c *RegistryBrokerClient) SearchErc8004ByAgentID(
	ctx context.Context,
	chainID int,
	agentID string,
	limit *int,
	page *int,
	sortBy string,
	sortOrder string,
) (JSONObject, error) {
	if chainID <= 0 {
		return nil, fmt.Errorf("chainID must be a positive integer")
	}
	trimmedAgentID := strings.TrimSpace(agentID)
	if trimmedAgentID == "" {
		return nil, ensureNonEmpty(agentID, "agentID")
	}

	nativeID := strconv.Itoa(chainID) + ":" + trimmedAgentID
	metadata := map[string][]any{
		"nativeId":   {nativeID},
		"networkKey": {"eip155:" + strconv.Itoa(chainID)},
	}

	query := SearchParams{
		Registries: []string{"erc-8004"},
		Metadata:   metadata,
		Limit:      1,
	}
	if limit != nil {
		query.Limit = *limit
	}
	if page != nil {
		query.Page = *page
	}
	query.SortBy = sortBy
	query.SortOrder = sortOrder

	return c.Search(ctx, query)
}

// Stats performs the requested operation.
func (c *RegistryBrokerClient) Stats(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/stats", nil, nil)
}

// Registries performs the requested operation.
func (c *RegistryBrokerClient) Registries(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/registries", nil, nil)
}

// GetAdditionalRegistries returns the requested value.
func (c *RegistryBrokerClient) GetAdditionalRegistries(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/register/additional-registries", nil, nil)
}

// PopularSearches performs the requested operation.
func (c *RegistryBrokerClient) PopularSearches(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/popular", nil, nil)
}

// ListProtocols performs the requested operation.
func (c *RegistryBrokerClient) ListProtocols(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/protocols", nil, nil)
}

// DetectProtocol performs the requested operation.
func (c *RegistryBrokerClient) DetectProtocol(
	ctx context.Context,
	message JSONObject,
) (JSONObject, error) {
	body := JSONObject{"message": message}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/detect-protocol",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// RegistrySearchByNamespace performs the requested operation.
func (c *RegistryBrokerClient) RegistrySearchByNamespace(
	ctx context.Context,
	registry string,
	query string,
) (JSONObject, error) {
	if err := ensureNonEmpty(registry, "registry"); err != nil {
		return nil, err
	}
	params := url.Values{}
	if strings.TrimSpace(query) != "" {
		params.Set("q", strings.TrimSpace(query))
	}
	path := "/registries/" + percentPath(registry) + "/search"
	return c.requestJSON(ctx, http.MethodGet, pathWithQuery(path, params), nil, nil)
}

// VectorSearch performs the requested operation.
func (c *RegistryBrokerClient) VectorSearch(
	ctx context.Context,
	request VectorSearchRequest,
) (JSONObject, error) {
	result, err := c.requestJSON(
		ctx,
		http.MethodPost,
		"/search",
		request,
		map[string]string{"content-type": "application/json"},
	)
	if err == nil {
		return result, nil
	}

	brokerErr, ok := err.(*RegistryBrokerError)
	if !ok || brokerErr.Status != http.StatusNotImplemented {
		return nil, err
	}

	fallbackParams := buildVectorFallbackSearchParams(request)
	fallback, fallbackErr := c.Search(ctx, fallbackParams)
	if fallbackErr != nil {
		return nil, fallbackErr
	}
	return convertSearchResultToVectorResponse(fallback), nil
}

// SearchStatus performs the requested operation.
func (c *RegistryBrokerClient) SearchStatus(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/search/status", nil, nil)
}

// WebsocketStats performs the requested operation.
func (c *RegistryBrokerClient) WebsocketStats(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/websocket/stats", nil, nil)
}

// MetricsSummary performs the requested operation.
func (c *RegistryBrokerClient) MetricsSummary(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/metrics", nil, nil)
}

// Facets performs the requested operation.
func (c *RegistryBrokerClient) Facets(ctx context.Context, adapter string) (JSONObject, error) {
	query := url.Values{}
	if strings.TrimSpace(adapter) != "" {
		query.Set("adapter", strings.TrimSpace(adapter))
	}
	return c.requestJSON(
		ctx,
		http.MethodGet,
		pathWithQuery("/search/facets", query),
		nil,
		nil,
	)
}

func buildSearchQuery(params SearchParams) string {
	query := url.Values{}
	addQueryString(query, "q", params.Q)
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	addQueryString(query, "registry", params.Registry)
	addQueryStrings(query, "registries", params.Registries)
	addQueryStrings(query, "capabilities", params.Capabilities)
	addQueryStrings(query, "protocols", params.Protocols)
	addQueryStrings(query, "adapters", params.Adapters)
	if params.MinTrust != nil {
		query.Set("minTrust", strconv.FormatFloat(*params.MinTrust, 'f', -1, 64))
	}
	if params.Metadata != nil {
		for key, values := range params.Metadata {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				continue
			}
			for _, value := range values {
				if value == nil {
					continue
				}
				query.Add("metadata."+trimmedKey, stringifyValue(value))
			}
		}
	}
	addQueryString(query, "type", params.Type)
	if params.Verified != nil && *params.Verified {
		query.Set("verified", "true")
	}
	if params.Online != nil && *params.Online {
		query.Set("online", "true")
	}
	addQueryString(query, "sortBy", params.SortBy)
	if strings.EqualFold(params.SortOrder, "asc") || strings.EqualFold(params.SortOrder, "desc") {
		query.Set("sortOrder", strings.ToLower(params.SortOrder))
	}

	if len(query) == 0 {
		return ""
	}
	return "?" + query.Encode()
}

func buildVectorFallbackSearchParams(request VectorSearchRequest) SearchParams {
	result := SearchParams{Q: request.Query}
	if request.Limit > 0 {
		result.Limit = request.Limit
	}
	if request.Offset > 0 {
		limit := request.Limit
		if limit <= 0 {
			limit = 20
		}
		result.Limit = limit
		result.Page = (request.Offset / limit) + 1
	}
	if request.Filter != nil {
		result.Registry = request.Filter.Registry
		result.Protocols = append(result.Protocols, request.Filter.Protocols...)
		result.Adapters = append(result.Adapters, request.Filter.Adapter...)
		result.Capabilities = append(result.Capabilities, request.Filter.Capabilities...)
		result.Type = request.Filter.Type
	}
	return result
}

func convertSearchResultToVectorResponse(result JSONObject) JSONObject {
	rawHits, _ := result["hits"].([]any)
	convertedHits := make([]any, 0, len(rawHits))
	for _, hit := range rawHits {
		convertedHits = append(convertedHits, JSONObject{
			"agent":      hit,
			"score":      0,
			"highlights": JSONObject{},
		})
	}
	total, _ := getNumberField(result, "total")
	limit, _ := getNumberField(result, "limit")
	page, _ := getNumberField(result, "page")
	totalVisible := page * limit
	limited := total > totalVisible || page > 1
	return JSONObject{
		"hits":           convertedHits,
		"total":          total,
		"took":           0,
		"totalAvailable": total,
		"visible":        len(convertedHits),
		"limited":        limited,
		"credits_used":   0,
	}
}

func stringifyValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}
