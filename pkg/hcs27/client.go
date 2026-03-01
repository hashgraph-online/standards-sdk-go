package hcs27

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient            *hedera.Client
	mirrorClient            *mirror.Client
	operatorID              hedera.AccountID
	operatorKey             hedera.PrivateKey
	network                 string
	inscriberAuthURL        string
	inscriberAPIURL         string
	publishMetadataOverride metadataPublisherFunc
}

type metadataPublisherFunc func(
	ctx context.Context,
	metadataBytes []byte,
) (string, *MetadataDigest, error)

const (
	inscriberWaitMaxAttempts = 120
	inscriberWaitIntervalMs  = 2000
	dataURLPartCount         = 2
)

type CreateTopicOptions struct {
	TTLSeconds          int64
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
}

// NewClient creates a new Client.
func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.OperatorAccountID) == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	if strings.TrimSpace(config.OperatorPrivateKey) == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	operatorID, err := hedera.AccountIDFromString(config.OperatorAccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	operatorKey, err := shared.ParsePrivateKey(config.OperatorPrivateKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(operatorID, operatorKey)

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
		APIKey:  config.MirrorAPIKey,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		hederaClient:     hederaClient,
		mirrorClient:     mirrorClient,
		operatorID:       operatorID,
		operatorKey:      operatorKey,
		network:          network,
		inscriberAuthURL: strings.TrimSpace(config.InscriberAuthURL),
		inscriberAPIURL:  strings.TrimSpace(config.InscriberAPIURL),
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// CreateCheckpointTopic creates the requested resource.
func (c *Client) CreateCheckpointTopic(
	ctx context.Context,
	options CreateTopicOptions,
) (string, string, error) {
	topicMemo := BuildTopicMemo(options.TTLSeconds)
	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(topicMemo)

	adminKey, err := c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin)
	if err != nil {
		return "", "", err
	}
	if adminKey != nil {
		transaction.SetAdminKey(*adminKey)
	}

	submitKey, err := c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit)
	if err != nil {
		return "", "", err
	}
	if submitKey != nil {
		transaction.SetSubmitKey(*submitKey)
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to create checkpoint topic: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to get create topic receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", "", fmt.Errorf("create topic receipt did not include topic ID")
	}

	return receipt.TopicID.String(), response.TransactionID.String(), nil
}

// PublishCheckpoint publishes the requested message payload.
func (c *Client) PublishCheckpoint(
	ctx context.Context,
	topicID string,
	metadata CheckpointMetadata,
	messageMemo string,
	transactionMemo string,
) (PublishResult, error) {
	if err := validateMetadata(metadata); err != nil {
		return PublishResult{}, err
	}
	if len(messageMemo) >= 300 {
		return PublishResult{}, fmt.Errorf("message memo must be less than 300 characters")
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return PublishResult{}, fmt.Errorf("failed to encode checkpoint metadata: %w", err)
	}

	message := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
		Memo:      messageMemo,
	}
	message, payload, inlineResolvedMetadata, err := c.prepareCheckpointPayload(ctx, message, metadataBytes)
	if err != nil {
		return PublishResult{}, err
	}

	resolver := c.ResolveHCS1Reference
	if len(inlineResolvedMetadata) > 0 {
		var metadataReference string
		if err := json.Unmarshal(message.Metadata, &metadataReference); err != nil {
			return PublishResult{}, fmt.Errorf("failed to decode metadata reference for validation: %w", err)
		}
		resolver = func(validationContext context.Context, hcs1Reference string) ([]byte, error) {
			if hcs1Reference == metadataReference {
				return inlineResolvedMetadata, nil
			}
			return c.ResolveHCS1Reference(validationContext, hcs1Reference)
		}
	}

	if _, err := ValidateCheckpointMessage(ctx, message, resolver); err != nil {
		return PublishResult{}, err
	}

	topic, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return PublishResult{}, fmt.Errorf("invalid topic ID %q: %w", topicID, err)
	}

	if transactionMemo == "" {
		transactionMemo = BuildTransactionMemo()
	}

	response, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topic).
		SetMessage(payload).
		SetTransactionMemo(transactionMemo).
		Execute(c.hederaClient)
	if err != nil {
		return PublishResult{}, fmt.Errorf("failed to publish checkpoint message: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return PublishResult{}, fmt.Errorf("failed to get checkpoint receipt: %w", err)
	}

	return PublishResult{
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}, nil
}

func (c *Client) prepareCheckpointPayload(
	ctx context.Context,
	message CheckpointMessage,
	metadataBytes []byte,
) (CheckpointMessage, []byte, []byte, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return CheckpointMessage{}, nil, nil, fmt.Errorf("failed to encode checkpoint message: %w", err)
	}
	if len(payload) <= 1024 {
		return message, payload, nil, nil
	}

	metadataReference, metadataDigest, err := c.publishMetadata(ctx, metadataBytes)
	if err != nil {
		return CheckpointMessage{}, nil, nil, err
	}

	referenceBytes, err := json.Marshal(metadataReference)
	if err != nil {
		return CheckpointMessage{}, nil, nil, fmt.Errorf("failed to encode metadata reference: %w", err)
	}
	message.Metadata = referenceBytes
	message.MetadataDigest = metadataDigest

	payload, err = json.Marshal(message)
	if err != nil {
		return CheckpointMessage{}, nil, nil, fmt.Errorf("failed to encode overflow checkpoint message: %w", err)
	}
	if len(payload) > 1024 {
		return CheckpointMessage{}, nil, nil, fmt.Errorf(
			"checkpoint overflow pointer message still exceeds 1024 bytes (got %d bytes)",
			len(payload),
		)
	}

	return message, payload, metadataBytes, nil
}

func (c *Client) publishMetadataHCS1(
	ctx context.Context,
	metadataBytes []byte,
) (string, *MetadataDigest, error) {
	network := inscriber.NetworkTestnet
	if strings.EqualFold(c.network, shared.NetworkMainnet) {
		network = inscriber.NetworkMainnet
	}

	authClient := inscriber.NewAuthClient(c.inscriberAuthURL)
	authResult, err := authClient.Authenticate(
		ctx,
		c.operatorID.String(),
		c.operatorKey.String(),
		network,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to authenticate inscriber client: %w", err)
	}

	inscriberClient, err := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: c.inscriberAPIURL,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create inscriber client: %w", err)
	}

	waitForConfirmation := true
	inscriptionResponse, err := inscriber.Inscribe(
		ctx,
		inscriber.InscriptionInput{
			Type:     inscriber.InscriptionInputTypeBuffer,
			Buffer:   metadataBytes,
			FileName: fmt.Sprintf("hcs27-checkpoint-%d.json", time.Now().UnixNano()),
			MimeType: "application/json",
		},
		inscriber.HederaClientConfig{
			AccountID:  c.operatorID.String(),
			PrivateKey: c.operatorKey.String(),
			Network:    network,
		},
		inscriber.InscriptionOptions{
			Mode:                inscriber.ModeFile,
			FileStandard:        "hcs-1",
			Network:             network,
			ConnectionMode:      inscriber.ConnectionModeWebSocket,
			WaitForConfirmation: &waitForConfirmation,
			WaitMaxAttempts:     inscriberWaitMaxAttempts,
			WaitInterval:        inscriberWaitIntervalMs,
		},
		inscriberClient,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to inscribe HCS-1 metadata: %w", err)
	}
	if !inscriptionResponse.Confirmed {
		return "", nil, fmt.Errorf("metadata inscription did not complete successfully")
	}

	inscriptionResult, ok := inscriptionResponse.Result.(inscriber.InscriptionResult)
	if !ok {
		return "", nil, fmt.Errorf("unexpected inscription result type %T", inscriptionResponse.Result)
	}

	inscribedTopicID := strings.TrimSpace(inscriptionResult.TopicID)
	if inscribedTopicID == "" && inscriptionResponse.Inscription != nil {
		inscribedTopicID = strings.TrimSpace(inscriptionResponse.Inscription.TopicID)
	}
	if inscribedTopicID == "" {
		return "", nil, fmt.Errorf("metadata inscription did not include topic ID")
	}

	reference := fmt.Sprintf("hcs://1/%s", inscribedTopicID)
	sum := sha256.Sum256(metadataBytes)
	digest := base64.RawURLEncoding.EncodeToString(sum[:])

	return reference, &MetadataDigest{
		Algorithm: "sha-256",
		DigestB64: digest,
	}, nil
}

func (c *Client) publishMetadata(
	ctx context.Context,
	metadataBytes []byte,
) (string, *MetadataDigest, error) {
	if c.publishMetadataOverride != nil {
		return c.publishMetadataOverride(ctx, metadataBytes)
	}
	return c.publishMetadataHCS1(ctx, metadataBytes)
}

// GetCheckpoints returns the requested value.
func (c *Client) GetCheckpoints(
	ctx context.Context,
	topicID string,
	resolver HCS1ResolverFunc,
) ([]CheckpointRecord, error) {
	effectiveResolver := resolver
	if effectiveResolver == nil {
		effectiveResolver = c.ResolveHCS1Reference
	}

	messages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Order: "asc",
	})
	if err != nil {
		return nil, err
	}

	records := make([]CheckpointRecord, 0, len(messages))
	for _, item := range messages {
		var message CheckpointMessage
		if err := mirror.DecodeMessageJSON(item, &message); err != nil {
			continue
		}

		metadata, err := ValidateCheckpointMessage(ctx, message, effectiveResolver)
		if err != nil {
			continue
		}

		record := CheckpointRecord{
			TopicID:            topicID,
			Sequence:           item.SequenceNumber,
			ConsensusTimestamp: item.ConsensusTimestamp,
			Payer:              item.PayerAccountID,
			Message:            message,
			EffectiveMetadata:  metadata,
		}
		records = append(records, record)
	}

	return records, nil
}

// ValidateCheckpointChain validates the provided input value.
func ValidateCheckpointChain(records []CheckpointRecord) error {
	type previousRecord struct {
		TreeSize    uint64
		RootHashB64 string
	}

	streams := map[string]previousRecord{}
	for _, record := range records {
		streamID := fmt.Sprintf(
			"%s::%s",
			record.EffectiveMetadata.Stream.Registry,
			record.EffectiveMetadata.Stream.LogID,
		)

		previous, exists := streams[streamID]
		current := previousRecord{
			TreeSize:    record.EffectiveMetadata.Root.TreeSize,
			RootHashB64: record.EffectiveMetadata.Root.RootHashB64,
		}

		if exists {
			if current.TreeSize < previous.TreeSize {
				return fmt.Errorf("tree size decreased for stream %s", streamID)
			}
			if record.EffectiveMetadata.Previous == nil {
				return fmt.Errorf("missing prev linkage for stream %s", streamID)
			}
			if record.EffectiveMetadata.Previous.TreeSize != previous.TreeSize {
				return fmt.Errorf("prev.treeSize mismatch for stream %s", streamID)
			}
			if record.EffectiveMetadata.Previous.RootHashB64 != previous.RootHashB64 {
				return fmt.Errorf("prev.rootHashB64u mismatch for stream %s", streamID)
			}
		}

		streams[streamID] = current
	}

	return nil
}

// ResolveHCS1Reference resolves the requested identifier data.
func (c *Client) ResolveHCS1Reference(ctx context.Context, hcs1Reference string) ([]byte, error) {
	trimmedReference := strings.TrimSpace(hcs1Reference)

	pattern := regexp.MustCompile(`^hcs://1/(\d+\.\d+\.\d+)$`)
	matches := pattern.FindStringSubmatch(trimmedReference)
	if len(matches) != 2 {
		return nil, fmt.Errorf("invalid HCS-1 reference %q", hcs1Reference)
	}
	topicID := matches[1]

	topicMessages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Order: "asc",
	})
	if err != nil {
		return nil, err
	}
	if len(topicMessages) == 0 {
		return nil, fmt.Errorf("no HCS-1 payload found at %s", hcs1Reference)
	}

	return decodeHCS1PayloadFromMessage(hcs1Reference, topicMessages[0], topicMessages)
}

func decodeHCS1PayloadFromMessage(
	hcs1Reference string,
	message mirror.TopicMessage,
	topicMessages []mirror.TopicMessage,
) ([]byte, error) {
	payload, err := mirror.DecodeMessageData(message)
	if err != nil {
		return nil, err
	}

	if message.ChunkInfo == nil || message.ChunkInfo.Total <= 1 {
		return normalizeHCS1Payload(payload)
	}

	chunkTransactionID := extractChunkTransactionID(message.ChunkInfo.InitialTransactionID)
	if chunkTransactionID == "" {
		return nil, fmt.Errorf("chunked HCS-1 payload at %s is missing initial transaction ID", hcs1Reference)
	}

	chunks := map[int][]byte{}
	for _, topicMessage := range topicMessages {
		if topicMessage.ChunkInfo == nil {
			continue
		}
		if topicMessage.ChunkInfo.Total != message.ChunkInfo.Total {
			continue
		}
		if extractChunkTransactionID(topicMessage.ChunkInfo.InitialTransactionID) != chunkTransactionID {
			continue
		}
		if topicMessage.ChunkInfo.Number <= 0 {
			continue
		}

		chunkPayload, decodeErr := mirror.DecodeMessageData(topicMessage)
		if decodeErr != nil {
			return nil, decodeErr
		}
		chunks[topicMessage.ChunkInfo.Number] = chunkPayload
	}

	if len(chunks) != message.ChunkInfo.Total {
		return nil, fmt.Errorf(
			"chunked HCS-1 payload at %s incomplete: expected %d chunks, found %d",
			hcs1Reference,
			message.ChunkInfo.Total,
			len(chunks),
		)
	}

	chunkNumbers := make([]int, 0, len(chunks))
	totalLength := 0
	for chunkNumber, chunkPayload := range chunks {
		chunkNumbers = append(chunkNumbers, chunkNumber)
		totalLength += len(chunkPayload)
	}
	sort.Ints(chunkNumbers)

	combined := make([]byte, 0, totalLength)
	for expected := 1; expected <= len(chunkNumbers); expected++ {
		if chunkNumbers[expected-1] != expected {
			return nil, fmt.Errorf(
				"chunked HCS-1 payload at %s missing chunk %d",
				hcs1Reference,
				expected,
			)
		}
		combined = append(combined, chunks[expected]...)
	}

	return normalizeHCS1Payload(combined)
}

func normalizeHCS1Payload(payload []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return payload, nil
	}
	if !bytes.Contains(trimmed, []byte(`"c"`)) {
		return payload, nil
	}

	wrappedContent := parseWrappedContent(trimmed)
	if wrappedContent == "" {
		return payload, nil
	}

	decodedContent, err := decodeDataURLPayload(wrappedContent)
	if err != nil {
		return nil, err
	}

	brotliReader := brotli.NewReader(bytes.NewReader(decodedContent))
	decompressed, err := io.ReadAll(brotliReader)
	if err == nil && len(decompressed) > 0 {
		return decompressed, nil
	}

	return decodedContent, nil
}

func parseWrappedContent(payload []byte) string {
	var wrapped struct {
		Content string `json:"c"`
	}
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return ""
	}
	return strings.TrimSpace(wrapped.Content)
}

func decodeDataURLPayload(input string) ([]byte, error) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "data:") {
		return nil, fmt.Errorf("unsupported wrapped HCS-1 payload format")
	}

	parts := strings.SplitN(trimmed, ",", dataURLPartCount)
	if len(parts) != dataURLPartCount {
		return nil, fmt.Errorf("invalid wrapped HCS-1 data URL")
	}

	header := strings.ToLower(parts[0])
	dataPart := parts[1]
	if strings.Contains(header, ";base64") {
		decoded, err := base64.StdEncoding.DecodeString(dataPart)
		if err != nil {
			return nil, fmt.Errorf("failed to decode wrapped HCS-1 base64 payload: %w", err)
		}
		return decoded, nil
	}

	unescaped, err := url.QueryUnescape(dataPart)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wrapped HCS-1 payload: %w", err)
	}
	return []byte(unescaped), nil
}

func extractChunkTransactionID(initialTransactionID any) string {
	switch typed := initialTransactionID.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		accountID, _ := typed["account_id"].(string)
		validStart, _ := typed["transaction_valid_start"].(string)
		if strings.TrimSpace(validStart) == "" {
			validStart, _ = typed["valid_start_timestamp"].(string)
		}
		if strings.TrimSpace(accountID) != "" && strings.TrimSpace(validStart) != "" {
			return accountID + "@" + validStart
		}
	case map[string]string:
		accountID := strings.TrimSpace(typed["account_id"])
		validStart := strings.TrimSpace(typed["transaction_valid_start"])
		if validStart == "" {
			validStart = strings.TrimSpace(typed["valid_start_timestamp"])
		}
		if accountID != "" && validStart != "" {
			return accountID + "@" + validStart
		}
	}

	return ""
}

func (c *Client) resolvePublicKey(rawKey string, useOperator bool) (*hedera.PublicKey, error) {
	if useOperator {
		publicKey := c.operatorKey.PublicKey()
		return &publicKey, nil
	}

	trimmed := strings.TrimSpace(rawKey)
	if trimmed == "" {
		return nil, nil
	}

	publicKey, publicErr := hedera.PublicKeyFromString(trimmed)
	if publicErr == nil {
		return &publicKey, nil
	}

	privateKey, privateErr := shared.ParsePrivateKey(trimmed)
	if privateErr != nil {
		return nil, fmt.Errorf("failed to parse key as public (%v) or private (%v)", publicErr, privateErr)
	}

	derivedKey := privateKey.PublicKey()
	return &derivedKey, nil
}
