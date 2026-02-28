package hcs7

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// BuildCreateRegistryTx builds and returns the configured value.
func BuildCreateRegistryTx(params CreateRegistryTxParams) *hedera.TopicCreateTransaction {
	ttl := params.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(BuildTopicMemo(ttl))
	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}
	return transaction
}

// BuildSubmitMessageTx builds and returns the configured value.
func BuildSubmitMessageTx(topicID string, message Message, transactionMemo string) (*hedera.TopicMessageSubmitTransaction, error) {
	if err := ValidateMessage(message); err != nil {
		return nil, err
	}
	parsedTopicID, err := hedera.TopicIDFromString(strings.TrimSpace(topicID))
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	payload, err := marshalMessage(message)
	if err != nil {
		return nil, err
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().SetTopicID(parsedTopicID).SetMessage(payload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	return transaction, nil
}

func marshalMessage(message Message) ([]byte, error) {
	type evmMessage struct {
		P      string           `json:"p"`
		Op     Operation        `json:"op"`
		Type   ConfigType       `json:"t"`
		Config EvmConfigPayload `json:"c"`
		Memo   string           `json:"m,omitempty"`
	}
	type wasmMessage struct {
		P      string            `json:"p"`
		Op     Operation         `json:"op"`
		Type   ConfigType        `json:"t"`
		Config WasmConfigPayload `json:"c"`
		Memo   string            `json:"m,omitempty"`
	}
	type metadataMessage struct {
		P       string         `json:"p"`
		Op      Operation      `json:"op"`
		TopicID string         `json:"t_id"`
		Data    map[string]any `json:"d"`
		Memo    string         `json:"m,omitempty"`
	}

	switch message.Op {
	case OperationRegisterConfig:
		if message.Type == ConfigTypeEVM {
			payload, ok := message.Config.(EvmConfigPayload)
			if !ok {
				return nil, fmt.Errorf("evm config payload is invalid")
			}
			return json.Marshal(evmMessage{
				P:      "hcs-7",
				Op:     OperationRegisterConfig,
				Type:   ConfigTypeEVM,
				Config: payload,
				Memo:   message.Memo,
			})
		}
		payload, ok := message.Config.(WasmConfigPayload)
		if !ok {
			return nil, fmt.Errorf("wasm config payload is invalid")
		}
		return json.Marshal(wasmMessage{
			P:      "hcs-7",
			Op:     OperationRegisterConfig,
			Type:   ConfigTypeWASM,
			Config: payload,
			Memo:   message.Memo,
		})
	case OperationRegister:
		return json.Marshal(metadataMessage{
			P:       "hcs-7",
			Op:      OperationRegister,
			TopicID: message.TopicID,
			Data:    message.Data,
			Memo:    message.Memo,
		})
	default:
		return nil, fmt.Errorf("unsupported operation %q", message.Op)
	}
}
