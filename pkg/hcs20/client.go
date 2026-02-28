package hcs20

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	network      string

	publicTopicID   string
	registryTopicID string
}

// NewClient creates a new HCS-20 SDK client.
func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	trimmedOperatorID := strings.TrimSpace(config.OperatorAccountID)
	if trimmedOperatorID == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	trimmedOperatorKey := strings.TrimSpace(config.OperatorPrivateKey)
	if trimmedOperatorKey == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	operatorID, err := hedera.AccountIDFromString(trimmedOperatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	operatorKey, err := shared.ParsePrivateKey(trimmedOperatorKey)
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

	publicTopicID := strings.TrimSpace(config.PublicTopicID)
	if publicTopicID == "" {
		publicTopicID = DefaultPublicTopicID
	}
	registryTopicID := strings.TrimSpace(config.RegistryTopicID)
	if registryTopicID == "" {
		registryTopicID = DefaultRegistryTopicID
	}

	return &Client{
		hederaClient:    hederaClient,
		mirrorClient:    mirrorClient,
		operatorID:      operatorID,
		operatorKey:     operatorKey,
		network:         network,
		publicTopicID:   publicTopicID,
		registryTopicID: registryTopicID,
	}, nil
}

// MirrorClient returns the configured mirror client.
func (client *Client) MirrorClient() *mirror.Client {
	return client.mirrorClient
}

// PublicTopicID returns the active HCS-20 public topic ID.
func (client *Client) PublicTopicID() string {
	return client.publicTopicID
}

// RegistryTopicID returns the active HCS-20 registry topic ID.
func (client *Client) RegistryTopicID() string {
	return client.registryTopicID
}

// SetPublicTopicID sets the default public topic ID used by mint/transfer/burn/deploy.
func (client *Client) SetPublicTopicID(topicID string) error {
	normalizedTopicID, err := NormalizeAccountID(topicID)
	if err != nil {
		return err
	}
	client.publicTopicID = normalizedTopicID
	return nil
}

// SetRegistryTopicID sets the default registry topic ID used by register.
func (client *Client) SetRegistryTopicID(topicID string) error {
	normalizedTopicID, err := NormalizeAccountID(topicID)
	if err != nil {
		return err
	}
	client.registryTopicID = normalizedTopicID
	return nil
}

// CreatePublicTopic creates a new topic and assigns it as the client's public topic ID.
func (client *Client) CreatePublicTopic(
	ctx context.Context,
	options CreateTopicOptions,
) (string, string, error) {
	topicID, transactionID, err := client.createTopic(ctx, options)
	if err != nil {
		return "", "", err
	}
	client.publicTopicID = topicID
	return topicID, transactionID, nil
}

// CreateRegistryTopic creates a new indexed HCS-2 registry and assigns it as the client's registry topic ID.
func (client *Client) CreateRegistryTopic(
	ctx context.Context,
	options hcs2.CreateRegistryOptions,
) (string, string, error) {
	hcs2Client, err := hcs2.NewClient(hcs2.ClientConfig{
		OperatorAccountID:  client.operatorID.String(),
		OperatorPrivateKey: client.operatorKey.String(),
		Network:            client.network,
		MirrorBaseURL:      client.mirrorClient.BaseURL(),
	})
	if err != nil {
		return "", "", err
	}

	if options.RegistryType != hcs2.RegistryTypeIndexed && options.RegistryType != hcs2.RegistryTypeNonIndexed {
		options.RegistryType = hcs2.RegistryTypeIndexed
	}
	if !options.UseOperatorAsAdmin && strings.TrimSpace(options.AdminKey) == "" {
		options.UseOperatorAsAdmin = true
	}
	if !options.UseOperatorAsSubmit && strings.TrimSpace(options.SubmitKey) == "" {
		options.UseOperatorAsSubmit = true
	}

	result, err := hcs2Client.CreateRegistry(ctx, options)
	if err != nil {
		return "", "", err
	}

	client.registryTopicID = result.TopicID
	return result.TopicID, result.TransactionID, nil
}

// DeployPoints deploys an HCS-20 points definition.
func (client *Client) DeployPoints(
	ctx context.Context,
	options DeployPointsOptions,
) (PointsInfo, error) {
	reportDeployProgress(options.ProgressCallback, DeployPointsProgress{
		Stage:      "creating-topic",
		Percentage: 20,
	})

	targetTopicID := client.publicTopicID
	if options.UsePrivateTopic {
		topicMemo := strings.TrimSpace(options.TopicMemo)
		if topicMemo == "" {
			topicMemo = fmt.Sprintf("hcs-20:%s", NormalizeTick(options.Tick))
		}

		privateTopicID, _, err := client.createTopic(ctx, CreateTopicOptions{
			Memo:                topicMemo,
			UseOperatorAsAdmin:  true,
			UseOperatorAsSubmit: true,
		})
		if err != nil {
			return PointsInfo{}, PointsDeploymentError{
				HCS20Error: HCS20Error{Message: fmt.Sprintf("failed to create private topic: %v", err)},
				Tick:       options.Tick,
			}
		}
		targetTopicID = privateTopicID
	}

	reportDeployProgress(options.ProgressCallback, DeployPointsProgress{
		Stage:      "submitting-deploy",
		Percentage: 50,
		TopicID:    targetTopicID,
	})

	submitTransaction, err := BuildHCS20DeployTx(DeployTxParams{
		TopicID:  targetTopicID,
		Name:     options.Name,
		Tick:     options.Tick,
		Max:      options.Max,
		Limit:    options.LimitPerMint,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	})
	if err != nil {
		return PointsInfo{}, err
	}

	result, err := client.submitTransaction(ctx, submitTransaction, targetTopicID)
	if err != nil {
		return PointsInfo{}, PointsDeploymentError{
			HCS20Error: HCS20Error{Message: fmt.Sprintf("failed to submit deploy message: %v", err)},
			Tick:       options.Tick,
		}
	}

	reportDeployProgress(options.ProgressCallback, DeployPointsProgress{
		Stage:      "confirming",
		Percentage: 80,
		TopicID:    targetTopicID,
		DeployTxID: result.TransactionID,
	})

	if !options.DisableMirrorCheck {
		if err := client.waitForMirrorSequence(ctx, targetTopicID, result.SequenceNumber, 15); err != nil {
			return PointsInfo{}, err
		}
	}

	reportDeployProgress(options.ProgressCallback, DeployPointsProgress{
		Stage:      "complete",
		Percentage: 100,
		TopicID:    targetTopicID,
		DeployTxID: result.TransactionID,
	})

	normalizedTick := NormalizeTick(options.Tick)
	maxSupply := strings.TrimSpace(options.Max)
	limitPerMint := strings.TrimSpace(options.LimitPerMint)
	metadata := strings.TrimSpace(options.Metadata)

	return PointsInfo{
		Name:                strings.TrimSpace(options.Name),
		Tick:                normalizedTick,
		MaxSupply:           maxSupply,
		LimitPerMint:        limitPerMint,
		Metadata:            metadata,
		TopicID:             targetTopicID,
		DeployerAccountID:   client.operatorID.String(),
		CurrentSupply:       "0",
		DeploymentTimestamp: result.ConsensusAt.UTC().Format(time.RFC3339Nano),
		IsPrivate:           options.UsePrivateTopic,
	}, nil
}

// MintPoints submits an HCS-20 mint message.
func (client *Client) MintPoints(
	ctx context.Context,
	options MintPointsOptions,
) (PointsTransaction, error) {
	reportMintProgress(options.ProgressCallback, MintPointsProgress{
		Stage:      "validating",
		Percentage: 20,
	})

	targetTopicID := client.resolveTopicID(options.TopicID)

	reportMintProgress(options.ProgressCallback, MintPointsProgress{
		Stage:      "submitting",
		Percentage: 50,
	})

	submitTransaction, err := BuildHCS20MintTx(MintTxParams{
		TopicID: targetTopicID,
		Tick:    options.Tick,
		Amount:  options.Amount,
		To:      options.To,
		Memo:    options.Memo,
	})
	if err != nil {
		return PointsTransaction{}, err
	}

	result, err := client.submitTransaction(ctx, submitTransaction, targetTopicID)
	if err != nil {
		return PointsTransaction{}, PointsMintError{
			HCS20Error:      HCS20Error{Message: fmt.Sprintf("failed to submit mint message: %v", err)},
			Tick:            options.Tick,
			RequestedAmount: options.Amount,
			AvailableSupply: "",
		}
	}

	reportMintProgress(options.ProgressCallback, MintPointsProgress{
		Stage:      "confirming",
		Percentage: 80,
		MintTxID:   result.TransactionID,
	})

	if !options.DisableMirrorCheck {
		if err := client.waitForMirrorSequence(ctx, targetTopicID, result.SequenceNumber, 15); err != nil {
			return PointsTransaction{}, err
		}
	}

	reportMintProgress(options.ProgressCallback, MintPointsProgress{
		Stage:      "complete",
		Percentage: 100,
		MintTxID:   result.TransactionID,
	})

	return PointsTransaction{
		ID:             result.TransactionID,
		Operation:      OperationMint,
		Tick:           NormalizeTick(options.Tick),
		Amount:         strings.TrimSpace(options.Amount),
		To:             strings.TrimSpace(options.To),
		Timestamp:      result.ConsensusAt.UTC().Format(time.RFC3339Nano),
		SequenceNumber: result.SequenceNumber,
		TopicID:        targetTopicID,
		TransactionID:  result.TransactionID,
		Memo:           strings.TrimSpace(options.Memo),
	}, nil
}

// TransferPoints submits an HCS-20 transfer message.
func (client *Client) TransferPoints(
	ctx context.Context,
	options TransferPointsOptions,
) (PointsTransaction, error) {
	reportTransferProgress(options.ProgressCallback, TransferPointsProgress{
		Stage:      "validating-balance",
		Percentage: 20,
	})

	targetTopicID := client.resolveTopicID(options.TopicID)
	fromAccountID, err := NormalizeAccountID(options.From)
	if err != nil {
		return PointsTransaction{}, err
	}
	toAccountID, err := NormalizeAccountID(options.To)
	if err != nil {
		return PointsTransaction{}, err
	}

	if targetTopicID == client.publicTopicID && fromAccountID != client.operatorID.String() {
		return PointsTransaction{}, PointsTransferError{
			HCS20Error: HCS20Error{Message: "for public topics, transaction payer must match sender"},
			Tick:       options.Tick,
			From:       fromAccountID,
			To:         toAccountID,
			Amount:     options.Amount,
		}
	}

	reportTransferProgress(options.ProgressCallback, TransferPointsProgress{
		Stage:      "submitting",
		Percentage: 50,
	})

	submitTransaction, err := BuildHCS20TransferTx(TransferTxParams{
		TopicID: targetTopicID,
		Tick:    options.Tick,
		Amount:  options.Amount,
		From:    fromAccountID,
		To:      toAccountID,
		Memo:    options.Memo,
	})
	if err != nil {
		return PointsTransaction{}, err
	}

	result, err := client.submitTransaction(ctx, submitTransaction, targetTopicID)
	if err != nil {
		return PointsTransaction{}, PointsTransferError{
			HCS20Error: HCS20Error{Message: fmt.Sprintf("failed to submit transfer message: %v", err)},
			Tick:       options.Tick,
			From:       fromAccountID,
			To:         toAccountID,
			Amount:     options.Amount,
		}
	}

	reportTransferProgress(options.ProgressCallback, TransferPointsProgress{
		Stage:        "confirming",
		Percentage:   80,
		TransferTxID: result.TransactionID,
	})

	if !options.DisableMirrorCheck {
		if err := client.waitForMirrorSequence(ctx, targetTopicID, result.SequenceNumber, 15); err != nil {
			return PointsTransaction{}, err
		}
	}

	reportTransferProgress(options.ProgressCallback, TransferPointsProgress{
		Stage:        "complete",
		Percentage:   100,
		TransferTxID: result.TransactionID,
	})

	return PointsTransaction{
		ID:             result.TransactionID,
		Operation:      OperationTransfer,
		Tick:           NormalizeTick(options.Tick),
		Amount:         strings.TrimSpace(options.Amount),
		From:           fromAccountID,
		To:             toAccountID,
		Timestamp:      result.ConsensusAt.UTC().Format(time.RFC3339Nano),
		SequenceNumber: result.SequenceNumber,
		TopicID:        targetTopicID,
		TransactionID:  result.TransactionID,
		Memo:           strings.TrimSpace(options.Memo),
	}, nil
}

// BurnPoints submits an HCS-20 burn message.
func (client *Client) BurnPoints(
	ctx context.Context,
	options BurnPointsOptions,
) (PointsTransaction, error) {
	reportBurnProgress(options.ProgressCallback, BurnPointsProgress{
		Stage:      "validating-balance",
		Percentage: 20,
	})

	targetTopicID := client.resolveTopicID(options.TopicID)
	fromAccountID, err := NormalizeAccountID(options.From)
	if err != nil {
		return PointsTransaction{}, err
	}

	if targetTopicID == client.publicTopicID && fromAccountID != client.operatorID.String() {
		return PointsTransaction{}, PointsBurnError{
			HCS20Error: HCS20Error{Message: "for public topics, transaction payer must match burner"},
			Tick:       options.Tick,
			From:       fromAccountID,
			Amount:     options.Amount,
		}
	}

	reportBurnProgress(options.ProgressCallback, BurnPointsProgress{
		Stage:      "submitting",
		Percentage: 50,
	})

	submitTransaction, err := BuildHCS20BurnTx(BurnTxParams{
		TopicID: targetTopicID,
		Tick:    options.Tick,
		Amount:  options.Amount,
		From:    fromAccountID,
		Memo:    options.Memo,
	})
	if err != nil {
		return PointsTransaction{}, err
	}

	result, err := client.submitTransaction(ctx, submitTransaction, targetTopicID)
	if err != nil {
		return PointsTransaction{}, PointsBurnError{
			HCS20Error: HCS20Error{Message: fmt.Sprintf("failed to submit burn message: %v", err)},
			Tick:       options.Tick,
			From:       fromAccountID,
			Amount:     options.Amount,
		}
	}

	reportBurnProgress(options.ProgressCallback, BurnPointsProgress{
		Stage:      "confirming",
		Percentage: 80,
		BurnTxID:   result.TransactionID,
	})

	if !options.DisableMirrorCheck {
		if err := client.waitForMirrorSequence(ctx, targetTopicID, result.SequenceNumber, 15); err != nil {
			return PointsTransaction{}, err
		}
	}

	reportBurnProgress(options.ProgressCallback, BurnPointsProgress{
		Stage:      "complete",
		Percentage: 100,
		BurnTxID:   result.TransactionID,
	})

	return PointsTransaction{
		ID:             result.TransactionID,
		Operation:      OperationBurn,
		Tick:           NormalizeTick(options.Tick),
		Amount:         strings.TrimSpace(options.Amount),
		From:           fromAccountID,
		Timestamp:      result.ConsensusAt.UTC().Format(time.RFC3339Nano),
		SequenceNumber: result.SequenceNumber,
		TopicID:        targetTopicID,
		TransactionID:  result.TransactionID,
		Memo:           strings.TrimSpace(options.Memo),
	}, nil
}

// RegisterTopic registers an HCS-20 topic on the configured registry topic.
func (client *Client) RegisterTopic(
	ctx context.Context,
	options RegisterTopicOptions,
) (OperationResult, error) {
	reportRegisterProgress(options.ProgressCallback, RegisterTopicProgress{
		Stage:      "validating",
		Percentage: 20,
	})

	registryTopicID := strings.TrimSpace(client.registryTopicID)
	if registryTopicID == "" {
		return OperationResult{}, TopicRegistrationError{
			HCS20Error: HCS20Error{Message: "registry topic ID is not configured"},
			TopicID:    "",
		}
	}

	reportRegisterProgress(options.ProgressCallback, RegisterTopicProgress{
		Stage:      "submitting",
		Percentage: 50,
	})

	submitTransaction, err := BuildHCS20RegisterTx(RegisterTxParams{
		RegistryTopicID: registryTopicID,
		Name:            options.Name,
		Metadata:        options.Metadata,
		IsPrivate:       options.IsPrivate,
		TopicID:         options.TopicID,
		Memo:            options.Memo,
	})
	if err != nil {
		return OperationResult{}, err
	}

	result, err := client.submitTransaction(ctx, submitTransaction, registryTopicID)
	if err != nil {
		return OperationResult{}, TopicRegistrationError{
			HCS20Error: HCS20Error{Message: fmt.Sprintf("failed to submit register message: %v", err)},
			TopicID:    options.TopicID,
		}
	}

	reportRegisterProgress(options.ProgressCallback, RegisterTopicProgress{
		Stage:        "confirming",
		Percentage:   80,
		RegisterTxID: result.TransactionID,
	})

	if !options.DisableMirrorCheck {
		if err := client.waitForMirrorSequence(ctx, registryTopicID, result.SequenceNumber, 15); err != nil {
			return OperationResult{}, err
		}
	}

	reportRegisterProgress(options.ProgressCallback, RegisterTopicProgress{
		Stage:        "complete",
		Percentage:   100,
		RegisterTxID: result.TransactionID,
	})

	return result, nil
}

func (client *Client) createTopic(
	ctx context.Context,
	options CreateTopicOptions,
) (string, string, error) {
	memo := strings.TrimSpace(options.Memo)
	if memo == "" {
		memo = "hcs-20"
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(memo)

	adminKey, err := client.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin)
	if err != nil {
		return "", "", err
	}
	if adminKey != nil {
		transaction.SetAdminKey(*adminKey)
	}

	submitKey, err := client.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit)
	if err != nil {
		return "", "", err
	}
	if submitKey != nil {
		transaction.SetSubmitKey(*submitKey)
	}

	response, err := transaction.Execute(client.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to create topic: %w", err)
	}

	receipt, err := response.GetReceipt(client.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to get topic receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", "", fmt.Errorf("topic receipt did not include topic ID")
	}

	return receipt.TopicID.String(), response.TransactionID.String(), nil
}

func (client *Client) resolvePublicKey(
	value string,
	useOperator bool,
) (*hedera.PublicKey, error) {
	if useOperator {
		key := client.operatorKey.PublicKey()
		return &key, nil
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	publicKey, err := hedera.PublicKeyFromString(trimmed)
	if err == nil {
		return &publicKey, nil
	}

	privateKey, privateKeyErr := shared.ParsePrivateKey(trimmed)
	if privateKeyErr == nil {
		key := privateKey.PublicKey()
		return &key, nil
	}

	return nil, fmt.Errorf("invalid key %q", trimmed)
}

func (client *Client) resolveTopicID(topicID string) string {
	trimmed := strings.TrimSpace(topicID)
	if trimmed == "" {
		return client.publicTopicID
	}
	return trimmed
}

func (client *Client) submitTransaction(
	ctx context.Context,
	transaction *hedera.TopicMessageSubmitTransaction,
	topicID string,
) (OperationResult, error) {
	response, err := transaction.Execute(client.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to execute topic message transaction: %w", err)
	}

	receipt, err := response.GetReceipt(client.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to get topic message receipt: %w", err)
	}

	record, recordErr := response.GetRecord(client.hederaClient)
	result := OperationResult{
		TopicID:        topicID,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}
	if recordErr == nil {
		result.ConsensusAt = record.ConsensusTimestamp
	}

	if result.ConsensusAt.IsZero() {
		result.ConsensusAt = time.Now().UTC()
	}

	return result, nil
}

func (client *Client) waitForMirrorSequence(
	ctx context.Context,
	topicID string,
	sequenceNumber int64,
	maxRetries int,
) error {
	if sequenceNumber <= 0 {
		return nil
	}
	if maxRetries <= 0 {
		maxRetries = 10
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		message, err := client.mirrorClient.GetTopicMessageBySequence(ctx, topicID, sequenceNumber)
		if err == nil && message != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return fmt.Errorf("mirror node did not return sequence %d for topic %s", sequenceNumber, topicID)
}

func reportDeployProgress(callback func(DeployPointsProgress), progress DeployPointsProgress) {
	if callback != nil {
		callback(progress)
	}
}

func reportMintProgress(callback func(MintPointsProgress), progress MintPointsProgress) {
	if callback != nil {
		callback(progress)
	}
}

func reportTransferProgress(callback func(TransferPointsProgress), progress TransferPointsProgress) {
	if callback != nil {
		callback(progress)
	}
}

func reportBurnProgress(callback func(BurnPointsProgress), progress BurnPointsProgress) {
	if callback != nil {
		callback(progress)
	}
}

func reportRegisterProgress(callback func(RegisterTopicProgress), progress RegisterTopicProgress) {
	if callback != nil {
		callback(progress)
	}
}
