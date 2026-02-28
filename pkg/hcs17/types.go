package hcs17

import (
	"fmt"
	"regexp"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type TopicState struct {
	TopicID           string
	LatestRunningHash string
}

type AccountStateInput struct {
	AccountID string
	PublicKey any
	Topics    []TopicState
}

type CompositeStateInput struct {
	CompositeAccountID            string
	CompositePublicKeyFingerprint string
	MemberStates                  []CompositeMemberState
	CompositeTopics               []TopicState
}

type CompositeMemberState struct {
	AccountID string
	StateHash string
}

type StateHashMessage struct {
	Protocol  string   `json:"p"`
	Operation string   `json:"op"`
	StateHash string   `json:"state_hash"`
	Topics    []string `json:"topics"`
	AccountID string   `json:"account_id"`
	Epoch     *int64   `json:"epoch,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
	Memo      string   `json:"m,omitempty"`
}

type StateHashResult struct {
	StateHash  string
	AccountID  string
	Timestamp  string
	TopicCount int
}

type CompositeStateHashResult struct {
	StateHash           string
	AccountID           string
	Timestamp           string
	TopicCount          int
	MemberCount         int
	CompositeTopicCount int
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateTopicOptions struct {
	TTLSeconds      int64
	AdminKey        hedera.Key
	SubmitKey       hedera.Key
	TransactionMemo string
}

type ComputeAndPublishOptions struct {
	AccountID        string
	AccountPublicKey any
	Topics           []string
	PublishTopicID   string
	Memo             string
}

type ComputeAndPublishResult struct {
	StateHash string
	Receipt   hedera.TransactionReceipt
}

type MessageRecord struct {
	Message            StateHashMessage
	ConsensusTimestamp string
	SequenceNumber     int64
	Payer              string
}

type TopicMemo struct {
	Type       HCS17TopicType
	TTLSeconds int64
}

type HCS17TopicType int

const (
	HCS17TopicTypeState HCS17TopicType = 0
)

func GenerateTopicMemo(ttlSeconds int64) string {
	if ttlSeconds <= 0 {
		ttlSeconds = 86400
	}
	return fmt.Sprintf("hcs-17:%d:%d", HCS17TopicTypeState, ttlSeconds)
}

func ParseTopicMemo(memo string) (*TopicMemo, error) {
	trimmed := strings.TrimSpace(memo)
	matches := regexp.MustCompile(`^hcs-17:(\d+):(\d+)$`).FindStringSubmatch(trimmed)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid HCS-17 memo format")
	}

	if matches[1] != "0" {
		return nil, fmt.Errorf("unsupported HCS-17 topic type")
	}

	var ttlSeconds int64
	_, err := fmt.Sscanf(matches[2], "%d", &ttlSeconds)
	if err != nil || ttlSeconds <= 0 {
		return nil, fmt.Errorf("invalid HCS-17 topic ttl")
	}

	return &TopicMemo{
		Type:       HCS17TopicTypeState,
		TTLSeconds: ttlSeconds,
	}, nil
}

func ValidateStateHashMessage(message StateHashMessage) []string {
	errors := make([]string, 0)
	if message.Protocol != "hcs-17" {
		errors = append(errors, "p must be hcs-17")
	}
	if message.Operation != "state_hash" {
		errors = append(errors, "op must be state_hash")
	}
	if strings.TrimSpace(message.StateHash) == "" {
		errors = append(errors, "state_hash is required")
	}
	if strings.TrimSpace(message.AccountID) == "" {
		errors = append(errors, "account_id is required")
	}
	if message.Topics == nil {
		errors = append(errors, "topics is required")
	}
	if message.Epoch != nil && *message.Epoch < 0 {
		errors = append(errors, "epoch must be non-negative")
	}
	return errors
}
