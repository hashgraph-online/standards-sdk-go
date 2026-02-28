package hcs2

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type CreateRegistryTxParams struct {
	RegistryType RegistryType
	TTL          int64
	AdminKey     hedera.Key
	SubmitKey    hedera.Key
	MemoOverride string
}

type RegisterTxParams struct {
	RegistryTopicID string
	TargetTopicID   string
	Metadata        string
	Memo            string
	AnalyticsMemo   string
	Protocol        string
}

type UpdateTxParams struct {
	RegistryTopicID string
	UID             string
	TargetTopicID   string
	Metadata        string
	Memo            string
	AnalyticsMemo   string
	Protocol        string
}

type DeleteTxParams struct {
	RegistryTopicID string
	UID             string
	Memo            string
	AnalyticsMemo   string
	Protocol        string
}

type MigrateTxParams struct {
	RegistryTopicID string
	TargetTopicID   string
	Metadata        string
	Memo            string
	AnalyticsMemo   string
	Protocol        string
}

// BuildHCS2CreateRegistryTx builds and returns the configured value.
func BuildHCS2CreateRegistryTx(params CreateRegistryTxParams) *hedera.TopicCreateTransaction {
	registryType := params.RegistryType
	if registryType != RegistryTypeIndexed && registryType != RegistryTypeNonIndexed {
		registryType = RegistryTypeIndexed
	}

	ttl := params.TTL
	if ttl <= 0 {
		ttl = 86400
	}

	memo := strings.TrimSpace(params.MemoOverride)
	if memo == "" {
		memo = BuildTopicMemo(registryType, ttl)
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(memo)

	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}

	return transaction
}

// BuildHCS2RegisterTx builds and returns the configured value.
func BuildHCS2RegisterTx(params RegisterTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	protocol := strings.TrimSpace(params.Protocol)
	if protocol == "" {
		protocol = "hcs-2"
	}

	message := Message{
		P:        protocol,
		Op:       OperationRegister,
		TopicID:  params.TargetTopicID,
		Metadata: params.Metadata,
		Memo:     params.Memo,
	}
	return buildHCS2MessageTx(params.RegistryTopicID, message, params.AnalyticsMemo)
}

// BuildHCS2UpdateTx builds and returns the configured value.
func BuildHCS2UpdateTx(params UpdateTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	protocol := strings.TrimSpace(params.Protocol)
	if protocol == "" {
		protocol = "hcs-2"
	}

	message := Message{
		P:        protocol,
		Op:       OperationUpdate,
		UID:      params.UID,
		TopicID:  params.TargetTopicID,
		Metadata: params.Metadata,
		Memo:     params.Memo,
	}
	return buildHCS2MessageTx(params.RegistryTopicID, message, params.AnalyticsMemo)
}

// BuildHCS2DeleteTx builds and returns the configured value.
func BuildHCS2DeleteTx(params DeleteTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	protocol := strings.TrimSpace(params.Protocol)
	if protocol == "" {
		protocol = "hcs-2"
	}

	message := Message{
		P:    protocol,
		Op:   OperationDelete,
		UID:  params.UID,
		Memo: params.Memo,
	}
	return buildHCS2MessageTx(params.RegistryTopicID, message, params.AnalyticsMemo)
}

// BuildHCS2MigrateTx builds and returns the configured value.
func BuildHCS2MigrateTx(params MigrateTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	protocol := strings.TrimSpace(params.Protocol)
	if protocol == "" {
		protocol = "hcs-2"
	}

	message := Message{
		P:        protocol,
		Op:       OperationMigrate,
		TopicID:  params.TargetTopicID,
		Metadata: params.Metadata,
		Memo:     params.Memo,
	}
	return buildHCS2MessageTx(params.RegistryTopicID, message, params.AnalyticsMemo)
}

func buildHCS2MessageTx(
	registryTopicID string,
	message Message,
	transactionMemo string,
) (*hedera.TopicMessageSubmitTransaction, error) {
	if err := ValidateMessage(message); err != nil {
		return nil, err
	}

	trimmedTopicID := strings.TrimSpace(registryTopicID)
	if trimmedTopicID == "" {
		return nil, fmt.Errorf("registry topic ID is required")
	}

	topicID, err := hedera.TopicIDFromString(trimmedTopicID)
	if err != nil {
		return nil, fmt.Errorf("invalid registry topic ID: %w", err)
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HCS-2 message: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicID).
		SetMessage(payload)

	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(transactionMemo)
	}

	return transaction, nil
}
