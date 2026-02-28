package hcs16

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

var hcs16OperationEnumByOperation = map[FloraOperation]int{
	FloraOperationFloraCreated: 0,
	FloraOperationTransaction:  1,
	FloraOperationStateUpdate:  2,
	FloraOperationJoinRequest:  3,
	FloraOperationJoinVote:     4,
	FloraOperationJoinAccepted: 5,
}

var hcs16TopicTypeByOperation = map[FloraOperation]FloraTopicType{
	FloraOperationFloraCreated: FloraTopicTypeCommunication,
	FloraOperationTransaction:  FloraTopicTypeTransaction,
	FloraOperationStateUpdate:  FloraTopicTypeState,
	FloraOperationJoinRequest:  FloraTopicTypeCommunication,
	FloraOperationJoinVote:     FloraTopicTypeCommunication,
	FloraOperationJoinAccepted: FloraTopicTypeState,
}

func BuildCreateFloraTopicTx(params CreateFloraTopicOptions) (*hedera.TopicCreateTransaction, error) {
	if strings.TrimSpace(params.FloraAccountID) == "" {
		return nil, fmt.Errorf("flora account ID is required")
	}
	if !isValidFloraTopicType(params.TopicType) {
		return nil, fmt.Errorf("invalid flora topic type %d", params.TopicType)
	}

	transaction := hedera.NewTopicCreateTransaction().
		SetTopicMemo(encodeFloraTopicMemo(params.FloraAccountID, params.TopicType)).
		SetTransactionMemo(normalizeMemo(params.TransactionMemo, encodeTopicCreateTransactionMemo(params.TopicType)))

	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}

	if strings.TrimSpace(params.AutoRenewAccount) != "" {
		autoRenewAccountID, err := hedera.AccountIDFromString(params.AutoRenewAccount)
		if err != nil {
			return nil, fmt.Errorf("invalid auto renew account ID: %w", err)
		}
		transaction.SetAutoRenewAccountID(autoRenewAccountID)
	}

	return transaction, nil
}

func BuildCreateTransactionTopicTx(
	config TransactionTopicConfig,
) (*hedera.TopicCreateTransaction, error) {
	if strings.TrimSpace(config.Memo) == "" {
		return nil, fmt.Errorf("memo is required")
	}
	if config.FeeScheduleKey != nil || len(config.CustomFees) > 0 || len(config.FeeExemptKeys) > 0 {
		return nil, fmt.Errorf("HIP-991 fee configuration is not supported by this hedera-sdk-go version")
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(strings.TrimSpace(config.Memo))
	if config.AdminKey != nil {
		transaction.SetAdminKey(config.AdminKey)
	}
	if config.SubmitKey != nil {
		transaction.SetSubmitKey(config.SubmitKey)
	}
	return transaction, nil
}

func BuildCreateFloraAccountTx(
	keyList *hedera.KeyList,
	initialBalanceHbar float64,
	maxAutomaticTokenAssociations int32,
	transactionMemo string,
) (*hedera.AccountCreateTransaction, error) {
	if keyList == nil {
		return nil, fmt.Errorf("key list is required")
	}

	initial := initialBalanceHbar
	if initial <= 0 {
		initial = 1
	}
	maxAssociations := maxAutomaticTokenAssociations
	if maxAssociations == 0 {
		maxAssociations = -1
	}

	transaction := hedera.NewAccountCreateTransaction().
		SetKey(keyList).
		SetInitialBalance(hedera.NewHbar(initial)).
		SetMaxAutomaticTokenAssociations(maxAssociations).
		SetTransactionMemo(normalizeMemo(transactionMemo, HCS16FloraAccountCreateTransactionMemo))

	return transaction, nil
}

func BuildScheduleAccountKeyUpdateTx(
	floraAccountID string,
	newKeyList *hedera.KeyList,
	transactionMemo string,
) (*hedera.ScheduleCreateTransaction, error) {
	if strings.TrimSpace(floraAccountID) == "" {
		return nil, fmt.Errorf("flora account ID is required")
	}
	if newKeyList == nil {
		return nil, fmt.Errorf("new key list is required")
	}

	accountID, err := hedera.AccountIDFromString(floraAccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid flora account ID: %w", err)
	}

	inner := hedera.NewAccountUpdateTransaction().
		SetAccountID(accountID).
		SetKey(newKeyList).
		SetTransactionMemo(normalizeMemo(transactionMemo, HCS16AccountKeyUpdateTransactionMemo))

	scheduleTransaction, err := hedera.NewScheduleCreateTransaction().SetScheduledTransaction(inner)
	if err != nil {
		return nil, fmt.Errorf("failed to set scheduled account update transaction: %w", err)
	}
	return scheduleTransaction, nil
}

func BuildScheduleTopicKeyUpdateTx(
	topicID string,
	adminKey hedera.Key,
	submitKey hedera.Key,
	transactionMemo string,
) (*hedera.ScheduleCreateTransaction, error) {
	if strings.TrimSpace(topicID) == "" {
		return nil, fmt.Errorf("topic ID is required")
	}

	parsedTopicID, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	inner := hedera.NewTopicUpdateTransaction().SetTopicID(parsedTopicID)
	if adminKey != nil {
		inner.SetAdminKey(adminKey)
	}
	if submitKey != nil {
		inner.SetSubmitKey(submitKey)
	}
	inner.SetTransactionMemo(normalizeMemo(transactionMemo, HCS16TopicKeyUpdateTransactionMemo))

	scheduleTransaction, err := hedera.NewScheduleCreateTransaction().SetScheduledTransaction(inner)
	if err != nil {
		return nil, fmt.Errorf("failed to set scheduled topic update transaction: %w", err)
	}
	return scheduleTransaction, nil
}

func BuildMessageTx(
	topicID string,
	operatorID string,
	operation FloraOperation,
	body map[string]any,
	transactionMemo string,
) (*hedera.TopicMessageSubmitTransaction, error) {
	if strings.TrimSpace(topicID) == "" {
		return nil, fmt.Errorf("topic ID is required")
	}
	if strings.TrimSpace(operatorID) == "" {
		return nil, fmt.Errorf("operator ID is required")
	}
	if !isValidFloraOperation(operation) {
		return nil, fmt.Errorf("invalid flora operation %q", operation)
	}

	parsedTopicID, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	payload := FloraMessage{
		"p":           "hcs-16",
		"op":          operation,
		"operator_id": operatorID,
	}
	for key, value := range body {
		payload[key] = value
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode flora message payload: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(encodedPayload).
		SetTransactionMemo(normalizeMemo(transactionMemo, encodeMessageTransactionMemo(operation)))

	return transaction, nil
}

func BuildFloraCreatedTx(
	topicID string,
	operatorID string,
	floraAccountID string,
	topics FloraTopics,
) (*hedera.TopicMessageSubmitTransaction, error) {
	return BuildMessageTx(topicID, operatorID, FloraOperationFloraCreated, map[string]any{
		"flora_account_id": floraAccountID,
		"topics": map[string]string{
			"communication": topics.Communication,
			"transaction":   topics.Transaction,
			"state":         topics.State,
		},
	}, "")
}

func BuildTransactionTx(
	topicID string,
	operatorID string,
	scheduleID string,
	data string,
) (*hedera.TopicMessageSubmitTransaction, error) {
	return BuildMessageTx(topicID, operatorID, FloraOperationTransaction, map[string]any{
		"schedule_id": scheduleID,
		"data":        data,
		"m":           data,
	}, "")
}

func BuildStateUpdateTx(
	topicID string,
	operatorID string,
	hash string,
	epoch *int64,
	accountID string,
	topics []string,
	memo string,
	transactionMemo string,
) (*hedera.TopicMessageSubmitTransaction, error) {
	if strings.TrimSpace(topicID) == "" {
		return nil, fmt.Errorf("topic ID is required")
	}
	if strings.TrimSpace(operatorID) == "" {
		return nil, fmt.Errorf("operator ID is required")
	}
	parsedTopicID, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	messageAccountID := strings.TrimSpace(accountID)
	if messageAccountID == "" {
		messageAccountID = operatorID
	}

	payload := map[string]any{
		"p":          "hcs-17",
		"op":         "state_hash",
		"state_hash": hash,
		"topics":     topics,
		"account_id": messageAccountID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
		"m":          memo,
	}
	if epoch != nil {
		payload["epoch"] = *epoch
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode state update payload: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(encodedPayload).
		SetTransactionMemo(normalizeMemo(transactionMemo, HCS17StateHashTransactionMemo))

	return transaction, nil
}

func BuildFloraJoinRequestTx(
	topicID string,
	operatorID string,
	accountID string,
	connectionRequestID int64,
	connectionTopicID string,
	connectionSequence int64,
) (*hedera.TopicMessageSubmitTransaction, error) {
	return BuildMessageTx(topicID, operatorID, FloraOperationJoinRequest, map[string]any{
		"account_id":            accountID,
		"connection_request_id": connectionRequestID,
		"connection_topic_id":   connectionTopicID,
		"connection_seq":        connectionSequence,
	}, "")
}

func BuildFloraJoinVoteTx(
	topicID string,
	operatorID string,
	accountID string,
	approve bool,
	connectionRequestID int64,
	connectionSequence int64,
) (*hedera.TopicMessageSubmitTransaction, error) {
	return BuildMessageTx(topicID, operatorID, FloraOperationJoinVote, map[string]any{
		"account_id":            accountID,
		"approve":               approve,
		"connection_request_id": connectionRequestID,
		"connection_seq":        connectionSequence,
	}, "")
}

func BuildFloraJoinAcceptedTx(
	topicID string,
	operatorID string,
	members []string,
	epoch *int64,
) (*hedera.TopicMessageSubmitTransaction, error) {
	body := map[string]any{
		"members": members,
	}
	if epoch != nil {
		body["epoch"] = *epoch
	}
	return BuildMessageTx(topicID, operatorID, FloraOperationJoinAccepted, body, "")
}

func normalizeMemo(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func encodeFloraTopicMemo(floraAccountID string, topicType FloraTopicType) string {
	return fmt.Sprintf("hcs-16:%s:%d", strings.TrimSpace(floraAccountID), topicType)
}

func encodeTopicCreateTransactionMemo(topicType FloraTopicType) string {
	return fmt.Sprintf(
		"hcs-16:op:%d:%d",
		hcs16OperationEnumByOperation[FloraOperationFloraCreated],
		topicType,
	)
}

func encodeMessageTransactionMemo(operation FloraOperation) string {
	return fmt.Sprintf(
		"hcs-16:op:%d:%d",
		hcs16OperationEnumByOperation[operation],
		hcs16TopicTypeByOperation[operation],
	)
}

func isValidFloraTopicType(topicType FloraTopicType) bool {
	return topicType == FloraTopicTypeCommunication ||
		topicType == FloraTopicTypeTransaction ||
		topicType == FloraTopicTypeState
}

func isValidFloraOperation(operation FloraOperation) bool {
	_, exists := hcs16OperationEnumByOperation[operation]
	return exists
}
