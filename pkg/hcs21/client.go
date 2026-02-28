package hcs21

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs2"
	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient *hedera.Client
	mirrorClient *mirror.Client
	operatorID   hedera.AccountID
	operatorKey  hedera.PrivateKey
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
		hederaClient: hederaClient,
		mirrorClient: mirrorClient,
		operatorID:   operatorID,
		operatorKey:  operatorKey,
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// BuildDeclaration performs the requested operation.
func (c *Client) BuildDeclaration(params BuildDeclarationParams) (AdapterDeclaration, error) {
	declaration := AdapterDeclaration{
		P:         Protocol,
		Op:        params.Op,
		AdapterID: strings.TrimSpace(params.AdapterID),
		Entity:    strings.TrimSpace(params.Entity),
		Package: AdapterPackage{
			Registry:  strings.TrimSpace(params.Package.Registry),
			Name:      strings.TrimSpace(params.Package.Name),
			Version:   strings.TrimSpace(params.Package.Version),
			Integrity: strings.TrimSpace(params.Package.Integrity),
		},
		Manifest:   strings.TrimSpace(params.Manifest),
		Config:     params.Config,
		StateModel: strings.TrimSpace(params.StateModel),
		Signature:  strings.TrimSpace(params.Signature),
	}
	if params.ManifestSequence > 0 {
		declaration.ManifestSequence = params.ManifestSequence
	}
	if err := ValidateDeclaration(declaration); err != nil {
		return AdapterDeclaration{}, err
	}
	return declaration, nil
}

// CreateRegistryTopic creates the requested resource.
func (c *Client) CreateRegistryTopic(
	ctx context.Context,
	options CreateRegistryTopicOptions,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	transaction, err := BuildCreateRegistryTopicTx(
		options.TTL,
		options.Indexed,
		options.Type,
		options.MetaTopicID,
		c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
	)
	if err != nil {
		return "", hedera.TransactionReceipt{}, err
	}
	if strings.TrimSpace(options.TransactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(options.TransactionMemo))
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to create HCS-21 topic")
	}
	return receipt.TopicID.String(), receipt, nil
}

// PublishDeclaration performs the requested operation.
func (c *Client) PublishDeclaration(
	ctx context.Context,
	options PublishDeclarationOptions,
) (hedera.TransactionReceipt, string, error) {
	_ = ctx

	transaction, err := BuildDeclarationMessageTx(options.TopicID, options.Declaration, options.TransactionMemo)
	if err != nil {
		return hedera.TransactionReceipt{}, "", err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to execute declaration transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to get declaration receipt: %w", err)
	}
	return receipt, response.TransactionID.String(), nil
}

// FetchDeclarations performs the requested operation.
func (c *Client) FetchDeclarations(
	ctx context.Context,
	topicID string,
	options FetchDeclarationsOptions,
) ([]AdapterDeclarationEnvelope, error) {
	order := strings.TrimSpace(options.Order)
	if order == "" {
		order = "asc"
	}
	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Limit: options.Limit,
		Order: order,
	})
	if err != nil {
		return nil, err
	}

	result := make([]AdapterDeclarationEnvelope, 0, len(items))
	for _, item := range items {
		declaration, err := decodeDeclaration(item.Message)
		if err != nil {
			continue
		}
		if err := ValidateDeclaration(declaration); err != nil {
			continue
		}
		result = append(result, AdapterDeclarationEnvelope{
			Declaration:        declaration,
			ConsensusTimestamp: item.ConsensusTimestamp,
			SequenceNumber:     item.SequenceNumber,
			Payer:              item.PayerAccountID,
		})
	}

	return result, nil
}

// CreateAdapterVersionPointerTopic creates the requested resource.
func (c *Client) CreateAdapterVersionPointerTopic(
	ctx context.Context,
	ttl int64,
	useOperatorAsAdmin bool,
	useOperatorAsSubmit bool,
	transactionMemo string,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	transaction := hcs2.BuildHCS2CreateRegistryTx(hcs2.CreateRegistryTxParams{
		RegistryType: hcs2.RegistryTypeNonIndexed,
		TTL:          ttl,
		AdminKey:     c.resolvePublicKey("", useOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey("", useOperatorAsSubmit),
	})
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute version pointer topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get version pointer topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to create HCS-2 version pointer topic")
	}
	return receipt.TopicID.String(), receipt, nil
}

// CreateRegistryDiscoveryTopic creates the requested resource.
func (c *Client) CreateRegistryDiscoveryTopic(
	ctx context.Context,
	ttl int64,
	useOperatorAsAdmin bool,
	useOperatorAsSubmit bool,
	transactionMemo string,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	transaction := hcs2.BuildHCS2CreateRegistryTx(hcs2.CreateRegistryTxParams{
		RegistryType: hcs2.RegistryTypeIndexed,
		TTL:          ttl,
		AdminKey:     c.resolvePublicKey("", useOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey("", useOperatorAsSubmit),
	})
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute discovery topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get discovery topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to create HCS-2 discovery topic")
	}
	return receipt.TopicID.String(), receipt, nil
}

// PublishVersionPointer performs the requested operation.
func (c *Client) PublishVersionPointer(
	ctx context.Context,
	versionTopicID string,
	declarationTopicID string,
	memo string,
	transactionMemo string,
) (hedera.TransactionReceipt, string, error) {
	_ = ctx

	transaction, err := hcs2.BuildHCS2RegisterTx(hcs2.RegisterTxParams{
		RegistryTopicID: versionTopicID,
		TargetTopicID:   declarationTopicID,
		Memo:            memo,
		AnalyticsMemo:   transactionMemo,
		Protocol:        "hcs-2",
	})
	if err != nil {
		return hedera.TransactionReceipt{}, "", err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to execute version pointer register transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to get version pointer register receipt: %w", err)
	}
	return receipt, response.TransactionID.String(), nil
}

// RegisterCategoryTopic performs the requested operation.
func (c *Client) RegisterCategoryTopic(
	ctx context.Context,
	discoveryTopicID string,
	categoryTopicID string,
	metadata string,
	memo string,
	transactionMemo string,
) (hedera.TransactionReceipt, string, error) {
	_ = ctx

	transaction, err := hcs2.BuildHCS2RegisterTx(hcs2.RegisterTxParams{
		RegistryTopicID: discoveryTopicID,
		TargetTopicID:   categoryTopicID,
		Metadata:        metadata,
		Memo:            memo,
		AnalyticsMemo:   transactionMemo,
		Protocol:        "hcs-2",
	})
	if err != nil {
		return hedera.TransactionReceipt{}, "", err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to execute category register transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to get category register receipt: %w", err)
	}
	return receipt, response.TransactionID.String(), nil
}

// PublishCategoryEntry performs the requested operation.
func (c *Client) PublishCategoryEntry(
	ctx context.Context,
	categoryTopicID string,
	adapterID string,
	versionTopicID string,
	metadata string,
	memo string,
	transactionMemo string,
) (hedera.TransactionReceipt, string, error) {
	_ = ctx

	resolvedMemo := strings.TrimSpace(memo)
	if resolvedMemo == "" {
		resolvedMemo = "adapter:" + strings.TrimSpace(adapterID)
	}

	transaction, err := hcs2.BuildHCS2RegisterTx(hcs2.RegisterTxParams{
		RegistryTopicID: categoryTopicID,
		TargetTopicID:   versionTopicID,
		Metadata:        metadata,
		Memo:            resolvedMemo,
		AnalyticsMemo:   transactionMemo,
		Protocol:        "hcs-2",
	})
	if err != nil {
		return hedera.TransactionReceipt{}, "", err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to execute category entry transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, "", fmt.Errorf("failed to get category entry receipt: %w", err)
	}
	return receipt, response.TransactionID.String(), nil
}

// ResolveLatestVersionPointer performs the requested operation.
func (c *Client) ResolveLatestVersionPointer(
	ctx context.Context,
	versionTopicID string,
) (string, int64, string, error) {
	registryClient, err := hcs2.NewClient(hcs2.ClientConfig{
		OperatorAccountID:  c.operatorID.String(),
		OperatorPrivateKey: c.operatorKey.String(),
		Network:            c.hederaClient.GetNetworkName(),
		MirrorBaseURL:      c.mirrorClient.BaseURL(),
	})
	if err != nil {
		return "", 0, "", err
	}

	registry, err := registryClient.GetRegistry(ctx, versionTopicID, hcs2.QueryRegistryOptions{
		Limit: 100,
		Order: "desc",
	})
	if err != nil {
		return "", 0, "", err
	}
	if len(registry.Entries) == 0 {
		return "", 0, "", fmt.Errorf("no version pointer messages found")
	}

	entry := registry.Entries[0]
	return entry.Message.TopicID, entry.Sequence, entry.Payer, nil
}

// FetchCategoryEntries performs the requested operation.
func (c *Client) FetchCategoryEntries(ctx context.Context, topicID string) ([]AdapterCategoryEntry, error) {
	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Order: "asc",
	})
	if err != nil {
		return nil, err
	}

	entries := make([]AdapterCategoryEntry, 0, len(items))
	for _, item := range items {
		message, err := decodeHCS2Message(item.Message)
		if err != nil {
			continue
		}
		if strings.TrimSpace(message.P) != "hcs-2" || message.Op != hcs2.OperationRegister {
			continue
		}
		adapterID := strings.TrimSpace(message.Memo)
		if strings.HasPrefix(adapterID, "adapter:") {
			adapterID = strings.TrimPrefix(adapterID, "adapter:")
		}
		if adapterID == "" {
			adapterID = strings.TrimSpace(message.TopicID)
		}
		entries = append(entries, AdapterCategoryEntry{
			AdapterID:          adapterID,
			AdapterTopicID:     strings.TrimSpace(message.TopicID),
			Metadata:           strings.TrimSpace(message.Metadata),
			Memo:               strings.TrimSpace(message.Memo),
			Payer:              item.PayerAccountID,
			SequenceNumber:     item.SequenceNumber,
			ConsensusTimestamp: item.ConsensusTimestamp,
		})
	}
	return entries, nil
}

func decodeDeclaration(encoded string) (AdapterDeclaration, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return AdapterDeclaration{}, err
	}
	var declaration AdapterDeclaration
	if err := json.Unmarshal(decoded, &declaration); err != nil {
		return AdapterDeclaration{}, err
	}
	return declaration, nil
}

func decodeHCS2Message(encoded string) (hcs2.Message, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return hcs2.Message{}, err
	}
	var message hcs2.Message
	if err := json.Unmarshal(decoded, &message); err != nil {
		return hcs2.Message{}, err
	}
	return message, nil
}

func (c *Client) resolvePublicKey(raw string, useOperator bool) hedera.Key {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" && !useOperator {
		return nil
	}
	if useOperator {
		return c.operatorKey.PublicKey()
	}
	publicKey, err := hedera.PublicKeyFromString(trimmed)
	if err != nil {
		return nil
	}
	return publicKey
}

