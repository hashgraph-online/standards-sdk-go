package hcs2

import (
	"fmt"
	"strconv"
	"strings"
)

type TopicMemo struct {
	RegistryType RegistryType
	TTL          int64
}

// BuildTopicMemo builds and returns the configured value.
func BuildTopicMemo(registryType RegistryType, ttl int64) string {
	return fmt.Sprintf("%s:%d:%d", defaultProtocol, registryType, ttl)
}

// ParseTopicMemo parses the provided input value.
func ParseTopicMemo(memo string) (*TopicMemo, bool) {
	parts := strings.Split(strings.TrimSpace(memo), ":")
	if len(parts) != 3 {
		return nil, false
	}
	if parts[0] != defaultProtocol {
		return nil, false
	}

	registryTypeValue, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, false
	}
	ttl, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, false
	}

	registryType := RegistryType(registryTypeValue)
	if registryType != RegistryTypeIndexed && registryType != RegistryTypeNonIndexed {
		return nil, false
	}

	return &TopicMemo{
		RegistryType: registryType,
		TTL:          ttl,
	}, true
}

// BuildTransactionMemo builds and returns the configured value.
func BuildTransactionMemo(operation Operation, registryType RegistryType) string {
	operationCode := map[Operation]int{
		OperationRegister: 0,
		OperationUpdate:   1,
		OperationDelete:   2,
		OperationMigrate:  3,
	}

	code := operationCode[operation]
	return fmt.Sprintf("%s:op:%d:%d", defaultProtocol, code, registryType)
}
