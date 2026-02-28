package hcs16

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashgraph-online/go-sdk/pkg/hcs11"
	"github.com/hashgraph-online/go-sdk/pkg/mirror"
	"github.com/hashgraph-online/go-sdk/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient     *hedera.Client
	mirrorClient     *mirror.Client
	operatorID       hedera.AccountID
	operatorKey      hedera.PrivateKey
	network          string
	inscriberAuthURL string
	inscriberAPIURL  string
}

func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	operatorID := strings.TrimSpace(config.OperatorAccountID)
	if operatorID == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	operatorKey := strings.TrimSpace(config.OperatorPrivateKey)
	if operatorKey == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	parsedOperatorID, err := hedera.AccountIDFromString(operatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	parsedOperatorKey, err := shared.ParsePrivateKey(operatorKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(parsedOperatorID, parsedOperatorKey)

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
		operatorID:       parsedOperatorID,
		operatorKey:      parsedOperatorKey,
		network:          network,
		inscriberAuthURL: strings.TrimSpace(config.InscriberAuthURL),
		inscriberAPIURL:  strings.TrimSpace(config.InscriberAPIURL),
	}, nil
}

func (c *Client) HederaClient() *hedera.Client {
	return c.hederaClient
}

func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

func (c *Client) ParseTopicMemo(memo string) *ParseTopicMemoResult {
	matches := regexp.MustCompile(`^hcs-16:([0-9.]+):(\d)$`).FindStringSubmatch(strings.TrimSpace(memo))
	if len(matches) != 3 {
		return nil
	}

	var topicType FloraTopicType
	switch matches[2] {
	case "0":
		topicType = FloraTopicTypeCommunication
	case "1":
		topicType = FloraTopicTypeTransaction
	case "2":
		topicType = FloraTopicTypeState
	default:
		return nil
	}

	return &ParseTopicMemoResult{
		Protocol:       "hcs-16",
		FloraAccountID: matches[1],
		TopicType:      topicType,
	}
}

func (c *Client) AssembleKeyList(ctx context.Context, members []string, threshold int) (*hedera.KeyList, error) {
	if len(members) == 0 {
		return nil, fmt.Errorf("members are required")
	}
	if threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive")
	}

	keys := make([]hedera.Key, 0, len(members))
	for _, memberAccountID := range members {
		publicKey, err := c.fetchAccountPublicKey(ctx, memberAccountID)
		if err != nil {
			return nil, err
		}
		keys = append(keys, publicKey)
	}

	keyList := hedera.KeyListWithThreshold(uint(threshold))
	keyList.AddAll(keys)
	return keyList, nil
}

func (c *Client) AssembleSubmitKeyList(ctx context.Context, members []string) (*hedera.KeyList, error) {
	if len(members) == 0 {
		return nil, fmt.Errorf("members are required")
	}

	keys := make([]hedera.Key, 0, len(members))
	for _, memberAccountID := range members {
		publicKey, err := c.fetchAccountPublicKey(ctx, memberAccountID)
		if err != nil {
			return nil, err
		}
		keys = append(keys, publicKey)
	}

	keyList := hedera.KeyListWithThreshold(1)
	keyList.AddAll(keys)
	return keyList, nil
}

func (c *Client) BuildFloraTopicCreateTxs(
	floraAccountID string,
	keyList *hedera.KeyList,
	submitKeyList *hedera.KeyList,
	autoRenewAccountID string,
) (map[FloraTopicType]*hedera.TopicCreateTransaction, error) {
	if keyList == nil {
		return nil, fmt.Errorf("key list is required")
	}
	if submitKeyList == nil {
		return nil, fmt.Errorf("submit key list is required")
	}

	communication, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID:   floraAccountID,
		TopicType:        FloraTopicTypeCommunication,
		AdminKey:         keyList,
		SubmitKey:        submitKeyList,
		AutoRenewAccount: autoRenewAccountID,
	})
	if err != nil {
		return nil, err
	}
	transaction, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID:   floraAccountID,
		TopicType:        FloraTopicTypeTransaction,
		AdminKey:         keyList,
		SubmitKey:        submitKeyList,
		AutoRenewAccount: autoRenewAccountID,
	})
	if err != nil {
		return nil, err
	}
	state, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID:   floraAccountID,
		TopicType:        FloraTopicTypeState,
		AdminKey:         keyList,
		SubmitKey:        submitKeyList,
		AutoRenewAccount: autoRenewAccountID,
	})
	if err != nil {
		return nil, err
	}

	return map[FloraTopicType]*hedera.TopicCreateTransaction{
		FloraTopicTypeCommunication: communication,
		FloraTopicTypeTransaction:   transaction,
		FloraTopicTypeState:         state,
	}, nil
}

func (c *Client) CreateFloraTopic(
	ctx context.Context,
	options CreateFloraTopicOptions,
) (string, error) {
	_ = ctx

	transaction, err := BuildCreateFloraTopicTx(options)
	if err != nil {
		return "", err
	}

	if len(options.SignerKeys) > 0 {
		frozen, freezeErr := transaction.FreezeWith(c.hederaClient)
		if freezeErr != nil {
			return "", fmt.Errorf("failed to freeze flora topic create transaction: %w", freezeErr)
		}
		for _, signerKey := range options.SignerKeys {
			frozen = frozen.Sign(signerKey)
		}
		response, executeErr := frozen.Execute(c.hederaClient)
		if executeErr != nil {
			return "", fmt.Errorf("failed to execute flora topic create transaction: %w", executeErr)
		}
		receipt, receiptErr := response.GetReceipt(c.hederaClient)
		if receiptErr != nil {
			return "", fmt.Errorf("failed to get flora topic create receipt: %w", receiptErr)
		}
		if receipt.TopicID == nil {
			return "", fmt.Errorf("failed to create flora topic")
		}
		return receipt.TopicID.String(), nil
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to execute flora topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to get flora topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", fmt.Errorf("failed to create flora topic")
	}
	return receipt.TopicID.String(), nil
}

func (c *Client) CreateFloraAccount(
	ctx context.Context,
	options CreateFloraAccountOptions,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	if options.KeyList == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("key list is required")
	}

	transaction, err := BuildCreateFloraAccountTx(
		options.KeyList,
		options.InitialBalanceHbar,
		options.MaxAutomaticTokenAssociations,
		"",
	)
	if err != nil {
		return "", hedera.TransactionReceipt{}, err
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute flora account create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get flora account create receipt: %w", err)
	}
	if receipt.AccountID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to create flora account")
	}

	return receipt.AccountID.String(), receipt, nil
}

func (c *Client) CreateFloraAccountWithTopics(
	ctx context.Context,
	options CreateFloraAccountWithTopicsOptions,
) (CreateFloraAccountWithTopicsResult, error) {
	keyList, err := c.AssembleKeyList(ctx, options.Members, options.Threshold)
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}
	submitKeyList, err := c.AssembleSubmitKeyList(ctx, options.Members)
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}

	initialBalance := options.InitialBalanceHbar
	if initialBalance <= 0 {
		initialBalance = 5
	}

	floraAccountID, _, err := c.CreateFloraAccount(ctx, CreateFloraAccountOptions{
		KeyList:                       keyList,
		InitialBalanceHbar:            initialBalance,
		MaxAutomaticTokenAssociations: -1,
	})
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}

	transactions, err := c.BuildFloraTopicCreateTxs(floraAccountID, keyList, submitKeyList, options.AutoRenewAccountID)
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}

	communicationTopicID, err := c.executeTopicCreate(transactions[FloraTopicTypeCommunication])
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}
	transactionTopicID, err := c.executeTopicCreate(transactions[FloraTopicTypeTransaction])
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}
	stateTopicID, err := c.executeTopicCreate(transactions[FloraTopicTypeState])
	if err != nil {
		return CreateFloraAccountWithTopicsResult{}, err
	}

	return CreateFloraAccountWithTopicsResult{
		FloraAccountID: floraAccountID,
		Topics: FloraTopics{
			Communication: communicationTopicID,
			Transaction:   transactionTopicID,
			State:         stateTopicID,
		},
	}, nil
}

func (c *Client) SendFloraCreated(
	ctx context.Context,
	topicID string,
	operatorID string,
	floraAccountID string,
	topics FloraTopics,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildFloraCreatedTx(topicID, operatorID, floraAccountID, topics)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	return c.executeMessage(transaction)
}

func (c *Client) SendTransaction(
	ctx context.Context,
	topicID string,
	operatorID string,
	scheduleID string,
	data string,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildTransactionTx(topicID, operatorID, scheduleID, data)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	return c.executeMessage(transaction)
}

func (c *Client) SendStateUpdate(
	ctx context.Context,
	topicID string,
	operatorID string,
	hash string,
	epoch *int64,
	accountID string,
	topics []string,
	memo string,
	transactionMemo string,
	signerKeys []hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildStateUpdateTx(
		topicID,
		operatorID,
		hash,
		epoch,
		accountID,
		topics,
		memo,
		transactionMemo,
	)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	return c.executeMessageWithSigners(transaction, signerKeys)
}

func (c *Client) SendFloraJoinRequest(
	ctx context.Context,
	topicID string,
	operatorID string,
	accountID string,
	connectionRequestID int64,
	connectionTopicID string,
	connectionSequence int64,
	signerKey *hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildFloraJoinRequestTx(
		topicID,
		operatorID,
		accountID,
		connectionRequestID,
		connectionTopicID,
		connectionSequence,
	)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	if signerKey == nil {
		return c.executeMessage(transaction)
	}
	return c.executeMessageWithSigners(transaction, []hedera.PrivateKey{*signerKey})
}

func (c *Client) SendFloraJoinVote(
	ctx context.Context,
	topicID string,
	operatorID string,
	accountID string,
	approve bool,
	connectionRequestID int64,
	connectionSequence int64,
	signerKey *hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildFloraJoinVoteTx(
		topicID,
		operatorID,
		accountID,
		approve,
		connectionRequestID,
		connectionSequence,
	)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	if signerKey == nil {
		return c.executeMessage(transaction)
	}
	return c.executeMessageWithSigners(transaction, []hedera.PrivateKey{*signerKey})
}

func (c *Client) SendFloraJoinAccepted(
	ctx context.Context,
	topicID string,
	operatorID string,
	members []string,
	epoch *int64,
	signerKeys []hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	_ = ctx
	transaction, err := BuildFloraJoinAcceptedTx(topicID, operatorID, members, epoch)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	return c.executeMessageWithSigners(transaction, signerKeys)
}

func (c *Client) SignSchedule(
	ctx context.Context,
	scheduleID string,
	signerKey hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	_ = ctx

	parsedScheduleID, err := hedera.ScheduleIDFromString(strings.TrimSpace(scheduleID))
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("invalid schedule ID: %w", err)
	}

	transaction := hedera.NewScheduleSignTransaction().SetScheduleID(parsedScheduleID)
	frozen, err := transaction.FreezeWith(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to freeze schedule sign transaction: %w", err)
	}

	signed := frozen.Sign(signerKey)
	response, err := signed.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to execute schedule sign transaction: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to get schedule sign receipt: %w", err)
	}
	return receipt, nil
}

func (c *Client) PublishFloraCreated(
	ctx context.Context,
	communicationTopicID string,
	operatorID string,
	floraAccountID string,
	topics FloraTopics,
) (hedera.TransactionReceipt, error) {
	return c.SendFloraCreated(ctx, communicationTopicID, operatorID, floraAccountID, topics)
}

func (c *Client) CreateFloraProfile(
	ctx context.Context,
	options CreateFloraProfileOptions,
) (CreateFloraProfileResult, error) {
	if strings.TrimSpace(options.DisplayName) == "" {
		return CreateFloraProfileResult{}, fmt.Errorf("display name is required")
	}

	profile := hcs11.HCS11Profile{
		Version:     "1.0",
		Type:        hcs11.ProfileTypeFlora,
		DisplayName: options.DisplayName,
		Members:     convertFloraMembers(options.Members),
		Threshold:   options.Threshold,
		Topics: &hcs11.FloraTopics{
			Communication: options.Topics.Communication,
			Transaction:   options.Topics.Transaction,
			State:         options.Topics.State,
		},
		InboundTopicID:  options.InboundTopicID,
		OutboundTopicID: options.OutboundTopicID,
		Bio:             options.Bio,
		Metadata:        options.Metadata,
		Policies:        options.Policies,
	}

	if strings.TrimSpace(profile.InboundTopicID) == "" {
		profile.InboundTopicID = options.Topics.Communication
	}
	if strings.TrimSpace(profile.OutboundTopicID) == "" {
		profile.OutboundTopicID = options.Topics.Transaction
	}

	hcs11Client, err := hcs11.NewClient(hcs11.ClientConfig{
		Network: c.network,
		Auth: hcs11.Auth{
			OperatorID: c.operatorID.String(),
			PrivateKey: c.operatorKey.String(),
		},
		MirrorBaseURL:    c.mirrorClient.BaseURL(),
		InscriberAuthURL: c.inscriberAuthURL,
		InscriberAPIURL:  c.inscriberAPIURL,
	})
	if err != nil {
		return CreateFloraProfileResult{}, err
	}

	inscription, err := hcs11Client.InscribeProfile(ctx, profile, hcs11.InscribeProfileOptions{
		WaitForConfirmation: true,
	})
	if err != nil {
		return CreateFloraProfileResult{}, err
	}
	if !inscription.Success {
		return CreateFloraProfileResult{}, fmt.Errorf("failed to inscribe flora profile: %s", inscription.Error)
	}

	updateResult, err := hcs11Client.UpdateAccountMemoWithProfile(
		ctx,
		options.FloraAccountID,
		inscription.ProfileTopicID,
	)
	if err != nil {
		return CreateFloraProfileResult{}, err
	}
	if !updateResult.Success {
		return CreateFloraProfileResult{}, fmt.Errorf("failed to update flora account memo: %s", updateResult.Error)
	}

	return CreateFloraProfileResult{
		ProfileTopicID: inscription.ProfileTopicID,
		TransactionID:  inscription.TransactionID,
	}, nil
}

func (c *Client) GetRecentMessages(
	ctx context.Context,
	topicID string,
	limit int,
	order string,
	opFilter FloraOperation,
) ([]FloraMessageRecord, error) {
	queryLimit := limit
	if queryLimit <= 0 {
		queryLimit = 25
	}
	queryOrder := strings.TrimSpace(order)
	if queryOrder == "" {
		queryOrder = "desc"
	}

	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Limit: queryLimit,
		Order: queryOrder,
	})
	if err != nil {
		return nil, err
	}

	results := make([]FloraMessageRecord, 0, len(items))
	for _, item := range items {
		decoded, decodeErr := base64.StdEncoding.DecodeString(item.Message)
		if decodeErr != nil {
			continue
		}
		var payload map[string]any
		if unmarshalErr := json.Unmarshal(decoded, &payload); unmarshalErr != nil {
			continue
		}
		protocol, _ := payload["p"].(string)
		if protocol != "hcs-16" {
			continue
		}
		if _, ok := payload["operator_id"].(string); !ok {
			continue
		}
		if opFilter != "" {
			if payloadOperation, _ := payload["op"].(string); payloadOperation != string(opFilter) {
				continue
			}
		}

		results = append(results, FloraMessageRecord{
			Message:            payload,
			ConsensusTimestamp: item.ConsensusTimestamp,
			SequenceNumber:     item.SequenceNumber,
			Payer:              item.PayerAccountID,
		})
	}

	return results, nil
}

func (c *Client) GetLatestMessage(
	ctx context.Context,
	topicID string,
	opFilter FloraOperation,
) (*FloraMessageRecord, error) {
	items, err := c.GetRecentMessages(ctx, topicID, 1, "desc", opFilter)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (c *Client) executeTopicCreate(
	transaction *hedera.TopicCreateTransaction,
) (string, error) {
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to execute flora topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to get flora topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", fmt.Errorf("failed to create flora topic")
	}
	return receipt.TopicID.String(), nil
}

func (c *Client) executeMessage(
	transaction *hedera.TopicMessageSubmitTransaction,
) (hedera.TransactionReceipt, error) {
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to execute flora message transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to get flora message receipt: %w", err)
	}
	return receipt, nil
}

func (c *Client) executeMessageWithSigners(
	transaction *hedera.TopicMessageSubmitTransaction,
	signerKeys []hedera.PrivateKey,
) (hedera.TransactionReceipt, error) {
	if len(signerKeys) == 0 {
		return c.executeMessage(transaction)
	}

	frozen, err := transaction.FreezeWith(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to freeze flora message transaction: %w", err)
	}
	for _, signerKey := range signerKeys {
		frozen = frozen.Sign(signerKey)
	}

	response, err := frozen.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to execute flora message transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to get flora message receipt: %w", err)
	}
	return receipt, nil
}

func (c *Client) fetchAccountPublicKey(ctx context.Context, accountID string) (hedera.PublicKey, error) {
	info, err := c.mirrorClient.GetAccount(ctx, accountID)
	if err != nil {
		return hedera.PublicKey{}, err
	}

	rawKey := extractMirrorKeyString(info.Key)
	if rawKey == "" {
		return hedera.PublicKey{}, fmt.Errorf("mirror node did not return a public key for account %s", accountID)
	}

	publicKey, err := hedera.PublicKeyFromString(strings.TrimSpace(rawKey))
	if err == nil {
		return publicKey, nil
	}

	ecdsaKey, ecdsaErr := hedera.PublicKeyFromStringECDSA(strings.TrimSpace(rawKey))
	if ecdsaErr == nil {
		return ecdsaKey, nil
	}

	edKey, edErr := hedera.PublicKeyFromStringEd25519(strings.TrimSpace(rawKey))
	if edErr == nil {
		return edKey, nil
	}

	return hedera.PublicKey{}, fmt.Errorf(
		"failed to parse account public key %q: generic=%v ecdsa=%v ed25519=%v",
		rawKey,
		err,
		ecdsaErr,
		edErr,
	)
}

func extractMirrorKeyString(raw map[string]any) string {
	if raw == nil {
		return ""
	}
	return extractMirrorKeyCandidate(raw)
}

func extractMirrorKeyCandidate(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		for _, key := range []string{"key", "ECDSA_secp256k1", "ed25519"} {
			if nested, ok := typed[key]; ok {
				candidate := extractMirrorKeyCandidate(nested)
				if candidate != "" {
					return candidate
				}
			}
		}
		for _, nested := range typed {
			switch nested.(type) {
			case map[string]any, []any:
				candidate := extractMirrorKeyCandidate(nested)
				if candidate != "" {
					return candidate
				}
			}
		}
	case []any:
		for _, nested := range typed {
			candidate := extractMirrorKeyCandidate(nested)
			if candidate != "" {
				return candidate
			}
		}
	}
	return ""
}

func convertFloraMembers(members []FloraMember) []hcs11.FloraMember {
	if len(members) == 0 {
		return nil
	}
	converted := make([]hcs11.FloraMember, 0, len(members))
	for _, member := range members {
		converted = append(converted, hcs11.FloraMember{
			AccountID: member.AccountID,
			PublicKey: member.PublicKey,
			Weight:    member.Weight,
		})
	}
	return converted
}
