package hcs2

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// maxPayloadBytes is the maximum HCS message payload size before overflow.
const maxPayloadBytes = 1024

// inscriberWaitMaxAttempts is the max number of polling attempts while waiting
// for an overflow inscription to complete.
const inscriberWaitMaxAttempts = 120

// inscriberWaitInterval is the polling interval between inscription status checks.
const inscriberWaitInterval = 2 * time.Second

// hcs1ReferencePattern matches an HCS-1 HRL like "hcs://1/0.0.12345".
var hcs1ReferencePattern = regexp.MustCompile(`^hcs://1/(\d+\.\d+\.\d+)$`)

// errNoPublicKey is returned when no public key is provided and the operator key is not used.
var errNoPublicKey = errors.New("no public key provided")

// Client is the HCS-2 SDK client.
type Client struct {
	hederaClient     *hedera.Client
	mirrorClient     *mirror.Client
	operatorID       hedera.AccountID
	operatorKey      hedera.PrivateKey
	network          string
	inscriberAuthURL string
	inscriberAPIURL  string
	registryTypeMap  map[string]RegistryType
	mutex            sync.RWMutex
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

	operatorID, err := hedera.AccountIDFromString(strings.TrimSpace(config.OperatorAccountID))
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
		registryTypeMap:  map[string]RegistryType{},
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// CreateRegistry creates the requested resource.
func (c *Client) CreateRegistry(
	ctx context.Context,
	options CreateRegistryOptions,
) (CreateRegistryResult, error) {
	registryType := options.RegistryType
	if registryType != RegistryTypeIndexed && registryType != RegistryTypeNonIndexed {
		registryType = RegistryTypeIndexed
	}
	ttl := options.TTL
	if ttl <= 0 {
		ttl = 86400
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(BuildTopicMemo(registryType, ttl))

	adminKey, err := c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin)
	if err != nil && !errors.Is(err, errNoPublicKey) {
		return CreateRegistryResult{}, err
	}
	if adminKey != nil {
		transaction.SetAdminKey(*adminKey)
	}

	submitKey, err := c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit)
	if err != nil && !errors.Is(err, errNoPublicKey) {
		return CreateRegistryResult{}, err
	}
	if submitKey != nil {
		transaction.SetSubmitKey(*submitKey)
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return CreateRegistryResult{}, fmt.Errorf("failed to execute create topic transaction: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return CreateRegistryResult{}, fmt.Errorf("failed to get create topic receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return CreateRegistryResult{}, fmt.Errorf("topic ID missing in create topic receipt")
	}

	topicID := receipt.TopicID.String()
	c.mutex.Lock()
	c.registryTypeMap[topicID] = registryType
	c.mutex.Unlock()

	return CreateRegistryResult{
		Success:       true,
		TopicID:       topicID,
		TransactionID: response.TransactionID.String(),
	}, nil
}

// RegisterEntry registers the requested resource.
func (c *Client) RegisterEntry(
	ctx context.Context,
	registryTopicID string,
	options RegisterEntryOptions,
	protocol string,
) (OperationResult, error) {
	if protocol == "" {
		protocol = defaultProtocol
	}
	message := Message{
		P:        protocol,
		Op:       OperationRegister,
		TopicID:  options.TargetTopicID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationRegister, registryType)
	}

	return c.submitMessage(ctx, registryTopicID, message, analyticsMemo)
}

// UpdateEntry updates the requested resource.
func (c *Client) UpdateEntry(
	ctx context.Context,
	registryTopicID string,
	options UpdateEntryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}
	if registryType != RegistryTypeIndexed {
		return OperationResult{}, fmt.Errorf("update is only valid for indexed registries")
	}

	message := Message{
		P:        defaultProtocol,
		Op:       OperationUpdate,
		TopicID:  options.TargetTopicID,
		UID:      options.UID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationUpdate, registryType)
	}

	return c.submitMessage(ctx, registryTopicID, message, analyticsMemo)
}

// DeleteEntry deletes the requested resource.
func (c *Client) DeleteEntry(
	ctx context.Context,
	registryTopicID string,
	options DeleteEntryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}
	if registryType != RegistryTypeIndexed {
		return OperationResult{}, fmt.Errorf("delete is only valid for indexed registries")
	}

	message := Message{
		P:    defaultProtocol,
		Op:   OperationDelete,
		UID:  options.UID,
		Memo: options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationDelete, registryType)
	}

	return c.submitMessage(ctx, registryTopicID, message, analyticsMemo)
}

// MigrateRegistry performs the requested operation.
func (c *Client) MigrateRegistry(
	ctx context.Context,
	registryTopicID string,
	options MigrateRegistryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}

	message := Message{
		P:        defaultProtocol,
		Op:       OperationMigrate,
		TopicID:  options.TargetTopicID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationMigrate, registryType)
	}

	return c.submitMessage(ctx, registryTopicID, message, analyticsMemo)
}

// GetRegistry returns the requested value.
func (c *Client) GetRegistry(
	ctx context.Context,
	topicID string,
	options QueryRegistryOptions,
) (TopicRegistry, error) {
	topicInfo, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return TopicRegistry{}, err
	}

	memoInfo, ok := ParseTopicMemo(topicInfo.Memo)
	if !ok {
		return TopicRegistry{}, fmt.Errorf("topic %s is not an HCS-2 registry", topicID)
	}

	order := options.Order
	if order == "" {
		order = "asc"
	}

	sequenceNumber := ""
	if options.Skip > 0 {
		sequenceNumber = fmt.Sprintf("gt:%d", options.Skip)
	}

	messages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: sequenceNumber,
		Limit:          options.Limit,
		Order:          order,
	})
	if err != nil {
		return TopicRegistry{}, err
	}

	entries := make([]RegistryEntry, 0, len(messages))
	var latestEntry *RegistryEntry

	for _, item := range messages {
		message, decodeErr := c.decodeRegistryMessage(ctx, item, options.ResolveOverflow)
		if decodeErr != nil {
			continue
		}
		if err := ValidateMessage(message); err != nil {
			continue
		}

		entry := RegistryEntry{
			TopicID:            topicID,
			Sequence:           item.SequenceNumber,
			Timestamp:          item.ConsensusTimestamp,
			Payer:              item.PayerAccountID,
			Message:            message,
			ConsensusTimestamp: item.ConsensusTimestamp,
			RegistryType:       memoInfo.RegistryType,
		}
		entries = append(entries, entry)

		if latestEntry == nil || entry.Timestamp > latestEntry.Timestamp {
			copyEntry := entry
			latestEntry = &copyEntry
		}
	}

	if memoInfo.RegistryType == RegistryTypeNonIndexed {
		if latestEntry == nil {
			entries = []RegistryEntry{}
		} else {
			entries = []RegistryEntry{*latestEntry}
		}
	}

	c.mutex.Lock()
	c.registryTypeMap[topicID] = memoInfo.RegistryType
	c.mutex.Unlock()

	return TopicRegistry{
		TopicID:      topicID,
		RegistryType: memoInfo.RegistryType,
		TTL:          memoInfo.TTL,
		Entries:      entries,
		LatestEntry:  latestEntry,
	}, nil
}

// GetTopicInfo returns the requested value.
func (c *Client) GetTopicInfo(ctx context.Context, topicID string) (mirror.TopicInfo, error) {
	return c.mirrorClient.GetTopicInfo(ctx, topicID)
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	registryTopicID string,
	payload Message,
	transactionMemo string,
) (OperationResult, error) {
	return c.submitMessage(ctx, registryTopicID, payload, transactionMemo)
}

// decodeRegistryMessage decodes a mirror node message into an HCS-2 Message.
// If the message is an overflow wrapper (contains data_ref), it optionally
// resolves the HCS-1 reference when resolveOverflow is true.
func (c *Client) decodeRegistryMessage(
	ctx context.Context,
	item mirror.TopicMessage,
	resolveOverflow bool,
) (Message, error) {
	var message Message
	if err := mirror.DecodeMessageJSON(item, &message); err == nil {
		return message, nil
	}

	// Try to decode as an overflow message.
	var overflow OverflowMessage
	if err := mirror.DecodeMessageJSON(item, &overflow); err != nil || overflow.DataRef == "" {
		return Message{}, fmt.Errorf("unable to decode message")
	}

	if !resolveOverflow {
		return Message{P: overflow.P, Op: overflow.Op}, nil
	}

	resolvedBytes, err := c.ResolveHCS1Reference(ctx, overflow.DataRef)
	if err != nil {
		return Message{}, fmt.Errorf("failed to resolve overflow: %w", err)
	}

	if err := json.Unmarshal(resolvedBytes, &message); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal resolved overflow: %w", err)
	}

	return message, nil
}

func (c *Client) submitMessage(
	ctx context.Context,
	registryTopicID string,
	message Message,
	transactionMemo string,
) (OperationResult, error) {
	topicID, err := hedera.TopicIDFromString(strings.TrimSpace(registryTopicID))
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid registry topic ID: %w", err)
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to marshal HCS-2 message: %w", err)
	}

	// If payload exceeds maxPayloadBytes, inscribe via HCS-1 and submit a reference.
	if len(payload) > maxPayloadBytes {
		hrl, digest, inscribeErr := c.inscribeOverflow(ctx, payload)
		if inscribeErr != nil {
			return OperationResult{}, fmt.Errorf("failed to inscribe overflow payload via HCS-1: %w", inscribeErr)
		}

		wrapper := OverflowMessage{
			P:             message.P,
			Op:            message.Op,
			DataRef:       hrl,
			DataRefDigest: digest,
		}
		payload, err = json.Marshal(wrapper)
		if err != nil {
			return OperationResult{}, fmt.Errorf("failed to marshal overflow wrapper: %w", err)
		}
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicID).
		SetMessage(payload)

	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(transactionMemo)
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to execute message submit transaction: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to get message submit receipt: %w", err)
	}

	return OperationResult{
		Success:        true,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber), //nolint:gosec // overflow won't occur in practice
	}, nil
}

// inscribeOverflow inscribes the payload via the Kiloscribe inscriber API and
// returns an HRL reference and SHA-256 base64url digest.
func (c *Client) inscribeOverflow(ctx context.Context, payload []byte) (hrl, digest string, err error) {
	network := inscriber.NetworkTestnet
	if strings.EqualFold(c.network, shared.NetworkMainnet) {
		network = inscriber.NetworkMainnet
	}

	authClient := inscriber.NewAuthClient(c.inscriberAuthURL)
	authResult, authErr := authClient.Authenticate(
		ctx,
		c.operatorID.String(),
		c.operatorKey.String(),
		network,
	)
	if authErr != nil {
		return "", "", fmt.Errorf("failed to authenticate inscriber client: %w", authErr)
	}

	inscriberClient, clientErr := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: c.inscriberAPIURL,
	})
	if clientErr != nil {
		return "", "", fmt.Errorf("failed to create inscriber client: %w", clientErr)
	}

	job, startErr := inscriberClient.StartInscription(ctx, inscriber.StartInscriptionRequest{
		HolderID:     c.operatorID.String(),
		Mode:         inscriber.ModeFile,
		Network:      network,
		FileStandard: "hcs-1",
		File: inscriber.FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString(payload),
			FileName: fmt.Sprintf("hcs2-overflow-%d.json", time.Now().UnixNano()),
			MimeType: "application/json",
		},
	})
	if startErr != nil {
		return "", "", fmt.Errorf("failed to start HCS-1 overflow inscription: %w", startErr)
	}
	if strings.TrimSpace(job.TransactionBytes) == "" {
		return "", "", fmt.Errorf("inscriber response did not include transaction bytes")
	}

	executedTxID, execErr := inscriber.ExecuteTransaction(
		ctx,
		job.TransactionBytes,
		inscriber.HederaClientConfig{
			AccountID:  c.operatorID.String(),
			PrivateKey: c.operatorKey.String(),
			Network:    network,
		},
	)
	if execErr != nil {
		return "", "", fmt.Errorf("failed to execute overflow inscription transaction: %w", execErr)
	}

	waited, waitErr := inscriberClient.WaitForInscription(ctx, executedTxID, inscriber.WaitOptions{
		MaxAttempts: inscriberWaitMaxAttempts,
		Interval:    inscriberWaitInterval,
	})
	if waitErr != nil {
		return "", "", fmt.Errorf("failed to wait for overflow inscription: %w", waitErr)
	}
	if !waited.Completed && !strings.EqualFold(waited.Status, "completed") {
		return "", "", fmt.Errorf("overflow inscription did not complete successfully")
	}

	inscribedTopicID := strings.TrimSpace(waited.TopicID)
	if inscribedTopicID == "" {
		inscribedTopicID = strings.TrimSpace(job.TopicID)
	}
	if inscribedTopicID == "" {
		return "", "", fmt.Errorf("overflow inscription did not return a topic ID")
	}

	hrl = fmt.Sprintf("hcs://1/%s", inscribedTopicID)
	sum := sha256.Sum256(payload)
	digest = base64.RawURLEncoding.EncodeToString(sum[:])

	return hrl, digest, nil
}

// ResolveHCS1Reference resolves an HCS-1 HRL (e.g. "hcs://1/0.0.12345") to the
// raw payload bytes stored on that topic.
func (c *Client) ResolveHCS1Reference(ctx context.Context, hcs1Reference string) ([]byte, error) {
	matches := hcs1ReferencePattern.FindStringSubmatch(strings.TrimSpace(hcs1Reference))
	if len(matches) != 2 { //nolint:mnd // regex capture group count
		return nil, fmt.Errorf("invalid HCS-1 reference %q", hcs1Reference)
	}
	topicID := matches[1]

	messages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Order: "asc",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HCS-1 payload from %s: %w", hcs1Reference, err)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("no HCS-1 payload found at %s", hcs1Reference)
	}

	// HCS-1 stores the payload as the message content of the first message.
	return base64.StdEncoding.DecodeString(messages[0].Message)
}

func (c *Client) resolveRegistryType(
	ctx context.Context,
	topicID string,
	override *RegistryType,
) (RegistryType, error) {
	if override != nil {
		return *override, nil
	}

	c.mutex.RLock()
	cachedType, ok := c.registryTypeMap[topicID]
	c.mutex.RUnlock()
	if ok {
		return cachedType, nil
	}

	topicInfo, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return RegistryTypeIndexed, err
	}
	memoInfo, parsed := ParseTopicMemo(topicInfo.Memo)
	if !parsed {
		return RegistryTypeIndexed, fmt.Errorf("topic %s is not an HCS-2 registry", topicID)
	}

	c.mutex.Lock()
	c.registryTypeMap[topicID] = memoInfo.RegistryType
	c.mutex.Unlock()

	return memoInfo.RegistryType, nil
}

func (c *Client) resolvePublicKey(rawKey string, useOperator bool) (*hedera.PublicKey, error) {
	if useOperator {
		publicKey := c.operatorKey.PublicKey()
		return &publicKey, nil
	}

	if strings.TrimSpace(rawKey) == "" {
		return nil, errNoPublicKey
	}

	publicKey, pubErr := hedera.PublicKeyFromString(rawKey)
	if pubErr == nil {
		return &publicKey, nil
	}

	privateKey, prvErr := shared.ParsePrivateKey(rawKey)
	if prvErr != nil {
		return nil, fmt.Errorf("failed to parse key as public (%w) or private (%w)", pubErr, prvErr)
	}

	derivedPublicKey := privateKey.PublicKey()
	return &derivedPublicKey, nil
}
