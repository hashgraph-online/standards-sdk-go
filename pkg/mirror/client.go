package mirror

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

type Config struct {
	Network    string
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Headers    map[string]string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	headers    map[string]string
}

type MessageQueryOptions struct {
	SequenceNumber string
	Limit          int
	Order          string
}

// NewClient creates a new Client.
func NewClient(config Config) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		if network == shared.NetworkMainnet {
			baseURL = "https://mainnet-public.mirrornode.hedera.com"
		} else {
			baseURL = "https://testnet.mirrornode.hedera.com"
		}
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid mirror base URL: %w", err)
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid mirror base URL: scheme must be http or https")
	}
	if strings.TrimSpace(parsedBaseURL.Host) == "" {
		return nil, fmt.Errorf("invalid mirror base URL: host is required")
	}
	baseURL = strings.TrimRight(parsedBaseURL.String(), "/")

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	headers := map[string]string{}
	for key, value := range config.Headers {
		headers[key] = value
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(config.APIKey),
		headers:    headers,
	}, nil
}

// BaseURL performs the requested operation.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// GetTopicInfo returns the requested value.
func (c *Client) GetTopicInfo(ctx context.Context, topicID string) (TopicInfo, error) {
	var topicInfo TopicInfo
	if strings.TrimSpace(topicID) == "" {
		return topicInfo, fmt.Errorf("topic ID is required")
	}

	path := fmt.Sprintf("/api/v1/topics/%s", topicID)
	if err := c.getJSON(ctx, path, &topicInfo); err != nil {
		return topicInfo, err
	}

	return topicInfo, nil
}

// GetAccount returns the requested value.
func (c *Client) GetAccount(ctx context.Context, accountID string) (AccountInfo, error) {
	var accountInfo AccountInfo
	normalizedAccountID := strings.TrimSpace(accountID)
	if normalizedAccountID == "" {
		return accountInfo, fmt.Errorf("account ID is required")
	}

	path := fmt.Sprintf("/api/v1/accounts/%s", normalizedAccountID)
	if err := c.getJSON(ctx, path, &accountInfo); err != nil {
		return accountInfo, err
	}

	return accountInfo, nil
}

// GetAccountMemo returns the requested value.
func (c *Client) GetAccountMemo(ctx context.Context, accountID string) (string, error) {
	accountInfo, err := c.GetAccount(ctx, accountID)
	if err != nil {
		return "", err
	}
	return accountInfo.Memo, nil
}

// GetTopicMessages returns the requested value.
func (c *Client) GetTopicMessages(
	ctx context.Context,
	topicID string,
	options MessageQueryOptions,
) ([]TopicMessage, error) {
	if strings.TrimSpace(topicID) == "" {
		return nil, fmt.Errorf("topic ID is required")
	}

	values := url.Values{}
	if options.SequenceNumber != "" {
		values.Set("sequencenumber", options.SequenceNumber)
	}
	if options.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", options.Limit))
	}
	if options.Order != "" {
		values.Set("order", options.Order)
	}

	endpoint := fmt.Sprintf("/api/v1/topics/%s/messages", topicID)
	if encoded := values.Encode(); encoded != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, encoded)
	}

	result := make([]TopicMessage, 0)
	next := endpoint

	for next != "" {
		var page topicMessagesResponse
		if err := c.getJSON(ctx, next, &page); err != nil {
			return nil, err
		}

		result = append(result, page.Messages...)
		next = page.Links.Next
	}

	return result, nil
}

// GetTopicMessageBySequence returns the requested value.
func (c *Client) GetTopicMessageBySequence(
	ctx context.Context,
	topicID string,
	sequence int64,
) (*TopicMessage, error) {
	if sequence <= 0 {
		return nil, fmt.Errorf("sequence must be positive")
	}

	messages, err := c.GetTopicMessages(ctx, topicID, MessageQueryOptions{
		SequenceNumber: fmt.Sprintf("eq:%d", sequence),
		Limit:          1,
		Order:          "asc",
	})
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil
	}

	return &messages[0], nil
}

// DecodeMessageData performs the requested operation.
func DecodeMessageData(message TopicMessage) ([]byte, error) {
	if strings.TrimSpace(message.Message) == "" {
		return nil, fmt.Errorf("message payload is empty")
	}
	return base64.StdEncoding.DecodeString(message.Message)
}

// DecodeMessageJSON performs the requested operation.
func DecodeMessageJSON[T any](message TopicMessage, target *T) error {
	payload, err := DecodeMessageData(message)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("failed to decode topic message JSON: %w", err)
	}

	return nil
}

// GetTransaction returns the requested value.
func (c *Client) GetTransaction(ctx context.Context, transactionID string) (*Transaction, error) {
	normalized := strings.TrimSpace(transactionID)
	if normalized == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	var response transactionsResponse
	path := fmt.Sprintf("/api/v1/transactions/%s", normalized)
	if err := c.getJSON(ctx, path, &response); err != nil {
		return nil, err
	}

	if len(response.Transactions) == 0 {
		return nil, nil
	}

	return &response.Transactions[0], nil
}

func (c *Client) getJSON(ctx context.Context, pathOrURL string, target any) error {
	requestURL := c.resolveURL(pathOrURL)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
	for key, value := range c.headers {
		request.Header.Set(key, value)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("mirror node request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read mirror node response: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"mirror node request failed with status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode mirror node response: %w", err)
	}

	return nil
}

func (c *Client) resolveURL(pathOrURL string) string {
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		return pathOrURL
	}

	path := pathOrURL
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return c.baseURL + path
}
