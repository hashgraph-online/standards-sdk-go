package inscriber

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Config struct {
	APIKey                       string
	Network                      Network
	BaseURL                      string
	HTTPClient                   *http.Client
	ConnectionMode               ConnectionMode
	WebSocketBaseURL             string
	WebSocketInactivityTimeoutMs int64
}

type Client struct {
	apiKey                       string
	network                      Network
	baseURL                      string
	httpClient                   *http.Client
	connectionMode               ConnectionMode
	webSocketBaseURL             string
	webSocketInactivityTimeoutMs int64
}

type WaitOptions struct {
	MaxAttempts int
	Interval    time.Duration
}

// NewClient creates a new Client.
func NewClient(config Config) (*Client, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	network := config.Network
	if network == "" {
		network = NetworkMainnet
	}
	if network != NetworkMainnet && network != NetworkTestnet {
		return nil, fmt.Errorf("network must be mainnet or testnet")
	}

	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = "https://v2-api.tier.bot/api"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}

	connectionMode := config.ConnectionMode
	if connectionMode == "" {
		connectionMode = ConnectionModeWebSocket
	}
	if connectionMode != ConnectionModeHTTP &&
		connectionMode != ConnectionModeWebSocket &&
		connectionMode != ConnectionModeAuto {
		return nil, fmt.Errorf("connection mode must be http, websocket, or auto")
	}

	return &Client{
		apiKey:                       apiKey,
		network:                      network,
		baseURL:                      baseURL,
		httpClient:                   httpClient,
		connectionMode:               connectionMode,
		webSocketBaseURL:             strings.TrimSpace(config.WebSocketBaseURL),
		webSocketInactivityTimeoutMs: config.WebSocketInactivityTimeoutMs,
	}, nil
}

// StartInscription performs the requested operation.
func (c *Client) StartInscription(
	ctx context.Context,
	request StartInscriptionRequest,
) (InscriptionJob, error) {
	if strings.TrimSpace(request.HolderID) == "" {
		return InscriptionJob{}, fmt.Errorf("holderId is required")
	}
	if request.Mode == "" {
		return InscriptionJob{}, fmt.Errorf("mode is required")
	}
	if request.File.Type != "url" && request.File.Type != "base64" {
		return InscriptionJob{}, fmt.Errorf("file.type must be url or base64")
	}

	body := map[string]any{
		"holderId": request.HolderID,
		"mode":     request.Mode,
		"network":  c.network,
	}

	if len(request.Metadata) > 0 {
		body["metadata"] = request.Metadata
	}
	if len(request.Tags) > 0 {
		body["tags"] = request.Tags
	}
	if request.ChunkSize > 0 {
		body["chunkSize"] = request.ChunkSize
	}
	if request.OnlyJSONCollection {
		body["onlyJSONCollection"] = boolToInt(request.OnlyJSONCollection)
	}
	if strings.TrimSpace(request.Creator) != "" {
		body["creator"] = request.Creator
	}
	if strings.TrimSpace(request.Description) != "" {
		body["description"] = request.Description
	}
	if strings.TrimSpace(request.FileStandard) != "" {
		body["fileStandard"] = request.FileStandard
	}
	if len(request.MetadataObject) > 0 {
		body["metadataObject"] = request.MetadataObject
	}
	if strings.TrimSpace(request.JSONFileURL) != "" {
		body["jsonFileURL"] = request.JSONFileURL
	}

	if request.File.Type == "url" {
		body["fileURL"] = request.File.URL
	} else {
		body["fileBase64"] = request.File.Base64
		body["fileName"] = request.File.FileName
		if request.File.MimeType != "" {
			body["fileMimeType"] = request.File.MimeType
		}
	}

	var raw map[string]any
	if err := c.postJSON(ctx, "/inscriptions/start-inscription", body, &raw); err != nil {
		return InscriptionJob{}, err
	}

	return parseInscriptionJob(raw)
}

// RetrieveInscription performs the requested operation.
func (c *Client) RetrieveInscription(ctx context.Context, txID string) (InscriptionJob, error) {
	normalizedID := normalizeTransactionID(txID)
	if normalizedID == "" {
		return InscriptionJob{}, fmt.Errorf("transaction ID is required")
	}

	endpoint := fmt.Sprintf("/inscriptions/retrieve-inscription?id=%s", url.QueryEscape(normalizedID))
	var raw map[string]any
	if err := c.getJSON(ctx, endpoint, &raw); err != nil {
		return InscriptionJob{}, err
	}

	job, err := parseInscriptionJob(raw)
	if err != nil {
		return InscriptionJob{}, err
	}
	if strings.EqualFold(job.Status, "completed") {
		job.Completed = true
	}
	if job.TxID == "" {
		job.TxID = normalizedID
	}

	return job, nil
}

// WaitForInscription performs the requested operation.
func (c *Client) WaitForInscription(
	ctx context.Context,
	txID string,
	options WaitOptions,
) (InscriptionJob, error) {
	maxAttempts := options.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 60
	}
	interval := options.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}

	var latest InscriptionJob
	for attempt := 0; attempt < maxAttempts; attempt++ {
		job, err := c.RetrieveInscription(ctx, txID)
		if err != nil {
			if isRetryableWaitError(err) && attempt < maxAttempts-1 {
				select {
				case <-ctx.Done():
					return InscriptionJob{}, ctx.Err()
				case <-time.After(interval):
				}
				continue
			}
			return InscriptionJob{}, err
		}
		latest = job

		if strings.EqualFold(job.Status, "failed") {
			if job.Error == "" {
				job.Error = "inscription failed"
			}
			return job, errors.New(job.Error)
		}
		if job.Completed || strings.EqualFold(job.Status, "completed") {
			return job, nil
		}

		select {
		case <-ctx.Done():
			return InscriptionJob{}, ctx.Err()
		case <-time.After(interval):
		}
	}

	return latest, fmt.Errorf("inscription did not complete within %d attempts", maxAttempts)
}

func isRetryableWaitError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "timed out") ||
		strings.Contains(lower, "temporarily unavailable") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "broken pipe") ||
		strings.Contains(lower, "eof")
}

// InscribeAndExecute inscribes the requested payload.
func (c *Client) InscribeAndExecute(
	ctx context.Context,
	request StartInscriptionRequest,
	clientConfig HederaClientConfig,
	waitForCompletion bool,
) (InscriptionResult, error) {
	job, err := c.StartInscription(ctx, request)
	if err != nil {
		return InscriptionResult{}, err
	}
	if strings.TrimSpace(job.TransactionBytes) == "" {
		return InscriptionResult{}, fmt.Errorf("inscription start did not include transaction bytes")
	}

	transactionID, err := ExecuteTransaction(ctx, job.TransactionBytes, clientConfig)
	if err != nil {
		return InscriptionResult{}, err
	}

	result := InscriptionResult{
		JobID:         normalizeTransactionID(job.TxID),
		TransactionID: normalizeTransactionID(transactionID),
		TopicID:       job.TopicID,
		Status:        job.Status,
		Completed:     false,
	}

	if !waitForCompletion {
		return result, nil
	}

	waited, err := c.WaitForInscription(ctx, transactionID, WaitOptions{})
	if err != nil {
		return InscriptionResult{}, err
	}

	result.TopicID = waited.TopicID
	result.Status = waited.Status
	result.Completed = waited.Completed

	return result, nil
}

// ExecuteTransaction performs the requested operation.
func ExecuteTransaction(
	ctx context.Context,
	transactionBytes string,
	config HederaClientConfig,
) (string, error) {
	network, err := shared.NormalizeNetwork(string(config.Network))
	if err != nil {
		return "", err
	}

	accountID, err := hedera.AccountIDFromString(strings.TrimSpace(config.AccountID))
	if err != nil {
		return "", fmt.Errorf("invalid account ID: %w", err)
	}
	privateKey, err := parseOperatorPrivateKey(ctx, network, accountID.String(), config.PrivateKey)
	if err != nil {
		return "", err
	}

	client, err := shared.NewHederaClient(network)
	if err != nil {
		return "", err
	}
	client.SetOperator(accountID, privateKey)

	rawBytes, err := base64.StdEncoding.DecodeString(transactionBytes)
	if err != nil {
		return "", fmt.Errorf("transaction bytes must be base64: %w", err)
	}

	type executeAttempt struct {
		operatorClient bool
		manualSign     bool
		label          string
	}

	attempts := []executeAttempt{
		{
			operatorClient: true,
			manualSign:     false,
			label:          "operator-auto-sign",
		},
		{
			operatorClient: true,
			manualSign:     true,
			label:          "operator-manual-sign",
		},
		{
			operatorClient: false,
			manualSign:     false,
			label:          "unsigned-pass-through",
		},
	}

	var invalidSignatureErrors []string
	for _, attempt := range attempts {
		executionClient, clientErr := shared.NewHederaClient(network)
		if clientErr != nil {
			return "", clientErr
		}
		if attempt.operatorClient {
			executionClient.SetOperator(accountID, privateKey)
		}

		transaction, decodeErr := hedera.TransactionFromBytes(rawBytes)
		if decodeErr != nil {
			return "", fmt.Errorf("failed to decode transaction bytes: %w", decodeErr)
		}

		executable := transaction
		if attempt.manualSign {
			signedTransaction, signErr := hedera.TransactionSign(transaction, privateKey)
			if signErr != nil {
				return "", fmt.Errorf("failed to sign transaction during %s: %w", attempt.label, signErr)
			}
			executable = signedTransaction
		}

		response, executeErr := hedera.TransactionExecute(executable, executionClient)
		if executeErr != nil {
			if strings.Contains(strings.ToUpper(executeErr.Error()), "INVALID_SIGNATURE") {
				invalidSignatureErrors = append(invalidSignatureErrors, fmt.Sprintf("%s=%v", attempt.label, executeErr))
				continue
			}
			return "", fmt.Errorf("failed to execute transaction via %s: %w", attempt.label, executeErr)
		}

		receipt, receiptErr := response.GetReceipt(executionClient)
		if receiptErr != nil {
			return "", fmt.Errorf("failed to get transaction receipt via %s: %w", attempt.label, receiptErr)
		}
		if receipt.Status.String() != "SUCCESS" {
			return "", fmt.Errorf("transaction via %s failed with status %s", attempt.label, receipt.Status.String())
		}

		return response.TransactionID.String(), nil
	}

	if len(invalidSignatureErrors) > 0 {
		rebuiltTransactionID, rebuildErr := executeRebuiltTransferTransaction(
			rawBytes,
			network,
			accountID,
			privateKey,
		)
		if rebuildErr == nil {
			return rebuiltTransactionID, nil
		}

		invalidSignatureErrors = append(invalidSignatureErrors, fmt.Sprintf("rebuilt-transfer=%v", rebuildErr))
		return "", fmt.Errorf("all execution strategies failed with INVALID_SIGNATURE: %s", strings.Join(invalidSignatureErrors, "; "))
	}

	return "", fmt.Errorf("no execution strategy succeeded")
}

func parseOperatorPrivateKey(
	ctx context.Context,
	network string,
	accountID string,
	rawPrivateKey string,
) (hedera.PrivateKey, error) {
	keyTypeHint, err := resolveMirrorKeyType(ctx, network, accountID)
	if err != nil {
		return hedera.PrivateKey{}, err
	}

	trimmedKey := strings.TrimSpace(rawPrivateKey)
	if trimmedKey == "" {
		return hedera.PrivateKey{}, fmt.Errorf("private key cannot be empty")
	}

	lowerHint := strings.ToLower(strings.TrimSpace(keyTypeHint))
	if strings.Contains(lowerHint, "ecdsa") {
		ecdsaKey, parseErr := hedera.PrivateKeyFromStringECDSA(trimmedKey)
		if parseErr == nil {
			return ecdsaKey, nil
		}
		return hedera.PrivateKey{}, fmt.Errorf("failed to parse private key as ECDSA for account %s: %w", accountID, parseErr)
	}
	if strings.Contains(lowerHint, "ed25519") {
		ed25519Key, parseErr := hedera.PrivateKeyFromStringEd25519(trimmedKey)
		if parseErr == nil {
			return ed25519Key, nil
		}
		return hedera.PrivateKey{}, fmt.Errorf("failed to parse private key as ED25519 for account %s: %w", accountID, parseErr)
	}

	return shared.ParsePrivateKey(trimmedKey)
}

func resolveMirrorKeyType(ctx context.Context, network string, accountID string) (string, error) {
	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
	})
	if err != nil {
		return "", err
	}

	accountInfo, err := mirrorClient.GetAccount(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch account key metadata from mirror node: %w", err)
	}

	typeValue, _ := accountInfo.Key["_type"].(string)
	return typeValue, nil
}

func executeRebuiltTransferTransaction(
	rawBytes []byte,
	network string,
	accountID hedera.AccountID,
	privateKey hedera.PrivateKey,
) (string, error) {
	decodedTransaction, err := hedera.TransactionFromBytes(rawBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decode transaction bytes for rebuild: %w", err)
	}

	transferTransaction, err := asTransferTransaction(decodedTransaction)
	if err != nil {
		return "", err
	}

	transactionID := transferTransaction.GetTransactionID()
	if transactionID.AccountID == nil || transactionID.AccountID.String() != accountID.String() {
		return "", fmt.Errorf(
			"cannot rebuild transfer transaction: payer account %v does not match operator account %s",
			transactionID.AccountID,
			accountID.String(),
		)
	}

	executionClient, err := shared.NewHederaClient(network)
	if err != nil {
		return "", err
	}
	executionClient.SetOperator(accountID, privateKey)

	rebuiltTransaction := hedera.NewTransferTransaction().
		SetTransactionID(transactionID).
		SetNodeAccountIDs(transferTransaction.GetNodeAccountIDs()).
		SetTransactionMemo(transferTransaction.GetTransactionMemo()).
		SetTransactionValidDuration(transferTransaction.GetTransactionValidDuration()).
		SetMaxTransactionFee(transferTransaction.GetMaxTransactionFee()).
		SetRegenerateTransactionID(false)

	for transferAccountID, transferAmount := range transferTransaction.GetHbarTransfers() {
		rebuiltTransaction.AddHbarTransfer(transferAccountID, transferAmount)
	}

	for tokenID, tokenTransfers := range transferTransaction.GetTokenTransfers() {
		for _, tokenTransfer := range tokenTransfers {
			if tokenTransfer.IsApproved {
				rebuiltTransaction.AddApprovedTokenTransfer(
					tokenID,
					tokenTransfer.AccountID,
					tokenTransfer.Amount,
					true,
				)
				continue
			}
			rebuiltTransaction.AddTokenTransfer(tokenID, tokenTransfer.AccountID, tokenTransfer.Amount)
		}
	}

	for tokenID, nftTransfers := range transferTransaction.GetNftTransfers() {
		for _, nftTransfer := range nftTransfers {
			nftID := hedera.NftID{
				TokenID:      tokenID,
				SerialNumber: nftTransfer.SerialNumber,
			}

			if nftTransfer.IsApproved {
				rebuiltTransaction.AddApprovedNftTransfer(
					nftID,
					nftTransfer.SenderAccountID,
					nftTransfer.ReceiverAccountID,
					true,
				)
				continue
			}

			rebuiltTransaction.AddNftTransfer(
				nftID,
				nftTransfer.SenderAccountID,
				nftTransfer.ReceiverAccountID,
			)
		}
	}

	if _, err := rebuiltTransaction.FreezeWith(executionClient); err != nil {
		return "", fmt.Errorf("failed to freeze rebuilt transfer transaction: %w", err)
	}

	rebuiltTransaction.Sign(privateKey)
	response, err := rebuiltTransaction.Execute(executionClient)
	if err != nil {
		return "", fmt.Errorf("failed to execute rebuilt transfer transaction: %w", err)
	}

	receipt, err := response.GetReceipt(executionClient)
	if err != nil {
		return "", fmt.Errorf("failed to fetch rebuilt transfer receipt: %w", err)
	}
	if receipt.Status.String() != "SUCCESS" {
		return "", fmt.Errorf("rebuilt transfer transaction failed with status %s", receipt.Status.String())
	}

	return response.TransactionID.String(), nil
}

func asTransferTransaction(transaction any) (*hedera.TransferTransaction, error) {
	switch typed := transaction.(type) {
	case *hedera.TransferTransaction:
		return typed, nil
	case hedera.TransferTransaction:
		value := typed
		return &value, nil
	default:
		return nil, fmt.Errorf("transaction bytes decoded to unsupported type %T (expected TransferTransaction)", transaction)
	}
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolveURL(endpoint), nil)
	if err != nil {
		return err
	}
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("inscriber API GET %s failed with status %d: %s", endpoint, response.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode inscriber API response: %w", err)
	}

	return nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.resolveURL(endpoint),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"inscriber API POST %s failed with status %d: %s",
			endpoint,
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	if err := json.Unmarshal(responseBody, target); err != nil {
		return fmt.Errorf("failed to decode inscriber API response: %w", err)
	}

	return nil
}

func (c *Client) resolveURL(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	if strings.HasPrefix(endpoint, "/") {
		return c.baseURL + endpoint
	}
	return c.baseURL + "/" + endpoint
}

func parseInscriptionJob(raw map[string]any) (InscriptionJob, error) {
	job := InscriptionJob{}

	if id, ok := raw["id"].(string); ok {
		job.ID = id
	}
	if status, ok := raw["status"].(string); ok {
		job.Status = status
	}
	if completed, ok := raw["completed"].(bool); ok {
		job.Completed = completed
	}
	if txID, ok := raw["tx_id"].(string); ok {
		job.TxID = txID
	}
	if topicID, ok := raw["topic_id"].(string); ok {
		job.TopicID = topicID
	}
	if transactionID, ok := raw["transactionId"].(string); ok {
		job.TransactionID = transactionID
	}
	if errorMessage, ok := raw["error"].(string); ok {
		job.Error = errorMessage
	}
	if totalCost, ok := raw["totalCost"].(float64); ok {
		job.TotalCost = int64(totalCost)
	}
	if totalMessages, ok := raw["totalMessages"].(float64); ok {
		job.TotalMessages = int64(totalMessages)
	}

	transactionBytes, err := normalizeTransactionBytes(raw["transactionBytes"])
	if err != nil {
		return InscriptionJob{}, err
	}
	job.TransactionBytes = transactionBytes

	return job, nil
}

func normalizeTransactionBytes(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case map[string]any:
		typeValue, _ := typed["type"].(string)
		if typeValue != "Buffer" {
			return "", fmt.Errorf("unsupported transactionBytes object type %q", typeValue)
		}
		items, ok := typed["data"].([]any)
		if !ok {
			return "", fmt.Errorf("transactionBytes Buffer object missing data array")
		}

		byteValues := make([]byte, 0, len(items))
		for _, item := range items {
			switch number := item.(type) {
			case float64:
				byteValues = append(byteValues, byte(number))
			case int:
				byteValues = append(byteValues, byte(number))
			default:
				return "", fmt.Errorf("transactionBytes data includes non-numeric value %T", item)
			}
		}

		return base64.StdEncoding.EncodeToString(byteValues), nil
	default:
		return "", fmt.Errorf("unsupported transactionBytes type %T", value)
	}
}

func normalizeTransactionID(txID string) string {
	trimmed := strings.TrimSpace(txID)
	if !strings.Contains(trimmed, "@") {
		return trimmed
	}

	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return trimmed
	}

	return parts[0] + "-" + strings.ReplaceAll(parts[1], ".", "-")
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
