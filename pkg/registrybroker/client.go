package registrybroker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBaseURL                = "https://hol.org/registry/api/v1"
	DefaultUserAgent              = "@hol-org/rb-client-go"
	DefaultProgressInterval       = 1500 * time.Millisecond
	DefaultProgressTimeout        = 5 * time.Minute
	DefaultHistoryTopUpHBAR       = 0.25
	MinimumRegistrationTopUpCount = 1
)

type RegistryBrokerClient struct {
	baseURL              string
	httpClient           *http.Client
	defaultHeaders       map[string]string
	registrationAutoTop  *AutoTopUpOptions
	historyAutoTop       *HistoryAutoTopUpOptions
	conversationContexts map[string][]ConversationContextState
	mutex                sync.RWMutex
}

// NewRegistryBrokerClient creates a new RegistryBrokerClient.
func NewRegistryBrokerClient(options RegistryBrokerClientOptions) (*RegistryBrokerClient, error) {
	timeout := options.HTTPTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	defaultHeaders := map[string]string{}
	for key, value := range options.DefaultHeaders {
		normalizedKey := normalizeHeaderName(key)
		trimmedValue := strings.TrimSpace(value)
		if normalizedKey != "" && trimmedValue != "" {
			defaultHeaders[normalizedKey] = trimmedValue
		}
	}

	if strings.TrimSpace(options.APIKey) != "" {
		defaultHeaders["x-api-key"] = strings.TrimSpace(options.APIKey)
	}
	if strings.TrimSpace(options.AccountID) != "" {
		defaultHeaders["x-account-id"] = strings.TrimSpace(options.AccountID)
	}
	if strings.TrimSpace(options.LedgerAPIKey) != "" && defaultHeaders["x-api-key"] == "" {
		defaultHeaders["x-api-key"] = strings.TrimSpace(options.LedgerAPIKey)
	}

	client := options.HTTPClient
	if client == nil {
		client = optionsClient(options.HTTPTimeout)
	}
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	return &RegistryBrokerClient{
		baseURL:              normalizeBaseURL(options.BaseURL),
		httpClient:           client,
		defaultHeaders:       defaultHeaders,
		registrationAutoTop:  cloneAutoTop(options.RegistrationAutoTop),
		historyAutoTop:       cloneHistoryAutoTop(options.HistoryAutoTop),
		conversationContexts: map[string][]ConversationContextState{},
	}, nil
}

func optionsClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		return nil
	}
	return &http.Client{Timeout: timeout}
}

// NewClient creates a new Client.
func NewClient(options RegistryBrokerClientOptions) (*RegistryBrokerClient, error) {
	return NewRegistryBrokerClient(options)
}

// BaseURL performs the requested operation.
func (c *RegistryBrokerClient) BaseURL() string {
	return c.baseURL
}

// SetAPIKey sets the requested value.
func (c *RegistryBrokerClient) SetAPIKey(apiKey string) {
	c.SetDefaultHeader("x-api-key", apiKey)
}

// SetLedgerAPIKey sets the requested value.
func (c *RegistryBrokerClient) SetLedgerAPIKey(apiKey string) {
	c.SetDefaultHeader("x-api-key", apiKey)
	c.mutex.Lock()
	delete(c.defaultHeaders, "x-ledger-api-key")
	c.mutex.Unlock()
}

// SetDefaultHeader sets the requested value.
func (c *RegistryBrokerClient) SetDefaultHeader(name string, value string) {
	normalizedName := normalizeHeaderName(name)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if normalizedName == "" || strings.TrimSpace(value) == "" {
		delete(c.defaultHeaders, normalizedName)
		return
	}
	c.defaultHeaders[normalizedName] = strings.TrimSpace(value)
}

// GetDefaultHeaders returns the requested value.
func (c *RegistryBrokerClient) GetDefaultHeaders() map[string]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	cloned := make(map[string]string, len(c.defaultHeaders))
	for key, value := range c.defaultHeaders {
		cloned[key] = value
	}
	return cloned
}

// BuildURL builds and returns the configured value.
func (c *RegistryBrokerClient) BuildURL(path string) string {
	normalizedPath := path
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}
	return c.baseURL + normalizedPath
}

func (c *RegistryBrokerClient) request(
	ctx context.Context,
	method string,
	path string,
	body any,
	headers map[string]string,
) ([]byte, http.Header, error) {
	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		requestBody = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.BuildURL(path), requestBody)
	if err != nil {
		return nil, nil, err
	}

	mergedHeaders := c.GetDefaultHeaders()
	for key, value := range headers {
		normalized := normalizeHeaderName(key)
		if normalized != "" && strings.TrimSpace(value) != "" {
			mergedHeaders[normalized] = strings.TrimSpace(value)
		}
	}
	if _, exists := mergedHeaders["accept"]; !exists {
		mergedHeaders["accept"] = "application/json"
	}
	if body != nil {
		if _, exists := mergedHeaders["content-type"]; !exists {
			mergedHeaders["content-type"] = "application/json"
		}
	}
	if _, exists := mergedHeaders["user-agent"]; !exists {
		mergedHeaders["user-agent"] = DefaultUserAgent
	}
	for key, value := range mergedHeaders {
		request.Header.Set(key, value)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		parsedBody := parseErrorResponseBody(response.Header, responseBody)
		return nil, nil, &RegistryBrokerError{
			Message:    "registry broker request failed",
			Status:     response.StatusCode,
			StatusText: response.Status,
			Body:       parsedBody,
		}
	}

	return responseBody, response.Header, nil
}

func (c *RegistryBrokerClient) requestJSON(
	ctx context.Context,
	method string,
	path string,
	body any,
	headers map[string]string,
) (JSONObject, error) {
	rawBody, rawHeaders, err := c.request(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	if !isJSONContentType(rawHeaders.Get("content-type")) {
		return nil, &RegistryBrokerParseError{
			Message: "expected JSON response from registry broker",
			Body:    strings.TrimSpace(string(rawBody)),
		}
	}
	var parsed JSONObject
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return nil, &RegistryBrokerParseError{
			Message: "failed to decode registry broker response",
			Body:    strings.TrimSpace(string(rawBody)),
			Cause:   err,
		}
	}
	return parsed, nil
}

func (c *RegistryBrokerClient) requestNoResponse(
	ctx context.Context,
	method string,
	path string,
	body any,
	headers map[string]string,
) error {
	_, _, err := c.request(ctx, method, path, body, headers)
	return err
}

func normalizeHeaderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeBaseURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = DefaultBaseURL
	}
	trimmed = strings.TrimRight(trimmed, "/")

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed == nil {
		return appendAPIVersion(trimmed)
	}

	hostname := strings.ToLower(parsed.Hostname())
	if hostname == "registry.hashgraphonline.com" || hostname == "hashgraphonline.com" {
		parsed.Host = strings.Replace(parsed.Host, parsed.Hostname(), "hol.org", 1)
	}
	if parsed.Hostname() == "hol.org" && !strings.HasPrefix(parsed.Path, "/registry") {
		if parsed.Path == "" || parsed.Path == "/" {
			parsed.Path = "/registry"
		} else {
			parsed.Path = "/registry" + parsed.Path
		}
	}
	return appendAPIVersion(strings.TrimRight(parsed.String(), "/"))
}

func appendAPIVersion(base string) string {
	if strings.HasSuffix(base, "/api/v1") || strings.HasSuffix(base, "/api/v2") {
		return base
	}
	if strings.HasSuffix(base, "/api") {
		return base + "/v1"
	}
	return base + "/api/v1"
}

func isJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

func parseErrorResponseBody(headers http.Header, body []byte) any {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}
	if isJSONContentType(headers.Get("content-type")) {
		var parsed any
		if err := json.Unmarshal(body, &parsed); err == nil {
			return parsed
		}
	}
	return trimmed
}

func cloneAutoTop(value *AutoTopUpOptions) *AutoTopUpOptions {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneHistoryAutoTop(value *HistoryAutoTopUpOptions) *HistoryAutoTopUpOptions {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func intPointerValue(value *int) (int, bool) {
	if value == nil {
		return 0, false
	}
	return *value, true
}

func boolPointerValue(value *bool) (bool, bool) {
	if value == nil {
		return false, false
	}
	return *value, true
}

func addQueryInt(values url.Values, key string, value *int) {
	if intValue, ok := intPointerValue(value); ok {
		values.Set(key, strconv.Itoa(intValue))
	}
}

func addQueryBool(values url.Values, key string, value *bool) {
	if boolValue, ok := boolPointerValue(value); ok {
		values.Set(key, strconv.FormatBool(boolValue))
	}
}

func addQueryString(values url.Values, key string, value string) {
	if strings.TrimSpace(value) != "" {
		values.Set(key, strings.TrimSpace(value))
	}
}

func addQueryStrings(values url.Values, key string, items []string) {
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			values.Add(key, trimmed)
		}
	}
}

func percentPath(value string) string {
	return url.PathEscape(strings.TrimSpace(value))
}

func pathWithQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}
	return path + "?" + query.Encode()
}

func bodyMap(value any) JSONObject {
	if value == nil {
		return JSONObject{}
	}
	if typed, ok := value.(JSONObject); ok {
		return typed
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return JSONObject{}
	}
	var parsed JSONObject
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return JSONObject{}
	}
	return parsed
}

func getNumberField(value JSONObject, key string) (float64, bool) {
	raw, exists := value[key]
	if !exists || raw == nil {
		return 0, false
	}
	switch typed := raw.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		if err == nil {
			return parsed, true
		}
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func (c *RegistryBrokerClient) extractInsufficientCreditsDetails(err error) (float64, bool) {
	brokerErr, ok := err.(*RegistryBrokerError)
	if !ok || brokerErr.Status != http.StatusPaymentRequired {
		return 0, false
	}
	bodyObject, ok := brokerErr.Body.(map[string]any)
	if !ok {
		return 0, false
	}
	shortfall, ok := getNumberField(bodyObject, "shortfallCredits")
	if !ok || shortfall <= 0 {
		return 0, false
	}
	return shortfall, true
}

func (c *RegistryBrokerClient) shouldAutoTopUpHistory(
	payload CreateSessionRequestPayload,
	err error,
) bool {
	if c.historyAutoTop == nil || payload.HistoryTTLSeconds == nil {
		return false
	}
	brokerErr, ok := err.(*RegistryBrokerError)
	if !ok || brokerErr.Status != http.StatusPaymentRequired {
		return false
	}
	body := brokerErr.Body
	message := ""
	if typed, ok := body.(string); ok {
		message = typed
	}
	if typed, ok := body.(map[string]any); ok {
		if typedMessage, ok := typed["message"].(string); ok {
			message = typedMessage
		} else if typedError, ok := typed["error"].(string); ok {
			message = typedError
		}
	}
	if strings.TrimSpace(message) == "" {
		return true
	}
	lowered := strings.ToLower(message)
	return strings.Contains(lowered, "history") || strings.Contains(lowered, "chat history")
}

func (c *RegistryBrokerClient) executeHistoryAutoTopUp(ctx context.Context, reason string) error {
	if c.historyAutoTop == nil {
		return nil
	}
	hbarAmount := c.historyAutoTop.HbarAmount
	if hbarAmount <= 0 {
		hbarAmount = DefaultHistoryTopUpHBAR
	}

	memo := strings.TrimSpace(c.historyAutoTop.Memo)
	if memo == "" {
		memo = "registry-broker-client:chat-history-topup"
	}
	_, err := c.PurchaseCreditsWithHbar(ctx, PurchaseCreditsWithHbarParams{
		AccountID:  c.historyAutoTop.AccountID,
		PrivateKey: c.historyAutoTop.PrivateKey,
		HbarAmount: hbarAmount,
		Memo:       memo,
		Metadata: JSONObject{
			"purpose": "chat-history",
			"reason":  reason,
		},
	})
	return err
}

func (c *RegistryBrokerClient) delay(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
