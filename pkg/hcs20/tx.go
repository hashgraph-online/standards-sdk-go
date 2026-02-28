package hcs20

import (
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type SubmitMessageTxParams struct {
	TopicID         string
	Payload         any
	TransactionMemo string
}

type DeployTxParams struct {
	TopicID         string
	Name            string
	Tick            string
	Max             string
	Limit           string
	Metadata        string
	Memo            string
	TransactionMemo string
}

type MintTxParams struct {
	TopicID         string
	Tick            string
	Amount          string
	To              string
	Memo            string
	TransactionMemo string
}

type BurnTxParams struct {
	TopicID         string
	Tick            string
	Amount          string
	From            string
	Memo            string
	TransactionMemo string
}

type TransferTxParams struct {
	TopicID         string
	Tick            string
	Amount          string
	From            string
	To              string
	Memo            string
	TransactionMemo string
}

type RegisterTxParams struct {
	RegistryTopicID string
	Name            string
	Metadata        string
	IsPrivate       bool
	TopicID         string
	Memo            string
	TransactionMemo string
}

// BuildHCS20SubmitMessageTx builds a generic HCS-20 submit transaction.
func BuildHCS20SubmitMessageTx(params SubmitMessageTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	trimmedTopicID := strings.TrimSpace(params.TopicID)
	if trimmedTopicID == "" {
		return nil, fmt.Errorf("topic ID is required")
	}

	topicID, err := hedera.TopicIDFromString(trimmedTopicID)
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	var payload []byte
	switch typedPayload := params.Payload.(type) {
	case []byte:
		payload = typedPayload
	case Message:
		encodedPayload, _, encodeErr := BuildMessagePayload(typedPayload)
		if encodeErr != nil {
			return nil, encodeErr
		}
		payload = encodedPayload
	default:
		return nil, fmt.Errorf("payload must be []byte or hcs20.Message")
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicID).
		SetMessage(payload)

	trimmedMemo := strings.TrimSpace(params.TransactionMemo)
	if trimmedMemo != "" {
		transaction.SetTransactionMemo(trimmedMemo)
	}

	return transaction, nil
}

// BuildHCS20DeployTx builds a deploy transaction.
func BuildHCS20DeployTx(params DeployTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	message := Message{
		Protocol:  ProtocolID,
		Operation: OperationDeploy,
		Name:      params.Name,
		Tick:      params.Tick,
		Max:       params.Max,
		Limit:     params.Limit,
		Metadata:  params.Metadata,
		Memo:      params.Memo,
	}
	return BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID:         params.TopicID,
		Payload:         message,
		TransactionMemo: params.TransactionMemo,
	})
}

// BuildHCS20MintTx builds a mint transaction.
func BuildHCS20MintTx(params MintTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	message := Message{
		Protocol:  ProtocolID,
		Operation: OperationMint,
		Tick:      params.Tick,
		Amount:    params.Amount,
		To:        params.To,
		Memo:      params.Memo,
	}
	return BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID:         params.TopicID,
		Payload:         message,
		TransactionMemo: params.TransactionMemo,
	})
}

// BuildHCS20BurnTx builds a burn transaction.
func BuildHCS20BurnTx(params BurnTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	message := Message{
		Protocol:  ProtocolID,
		Operation: OperationBurn,
		Tick:      params.Tick,
		Amount:    params.Amount,
		From:      params.From,
		Memo:      params.Memo,
	}
	return BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID:         params.TopicID,
		Payload:         message,
		TransactionMemo: params.TransactionMemo,
	})
}

// BuildHCS20TransferTx builds a transfer transaction.
func BuildHCS20TransferTx(params TransferTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	message := Message{
		Protocol:  ProtocolID,
		Operation: OperationTransfer,
		Tick:      params.Tick,
		Amount:    params.Amount,
		From:      params.From,
		To:        params.To,
		Memo:      params.Memo,
	}
	return BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID:         params.TopicID,
		Payload:         message,
		TransactionMemo: params.TransactionMemo,
	})
}

// BuildHCS20RegisterTx builds a registry registration transaction.
func BuildHCS20RegisterTx(params RegisterTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	isPrivate := params.IsPrivate
	message := Message{
		Protocol:  ProtocolID,
		Operation: OperationRegister,
		Name:      params.Name,
		Metadata:  params.Metadata,
		Private:   &isPrivate,
		TopicID:   params.TopicID,
		Memo:      params.Memo,
	}
	return BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID:         params.RegistryTopicID,
		Payload:         message,
		TransactionMemo: params.TransactionMemo,
	})
}
