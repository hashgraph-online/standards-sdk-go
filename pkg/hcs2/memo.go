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

func BuildTopicMemo(registryType RegistryType, ttl int64) string {
	return fmt.Sprintf("hcs-2:%d:%d", registryType, ttl)
}

func ParseTopicMemo(memo string) (*TopicMemo, bool) {
	parts := strings.Split(strings.TrimSpace(memo), ":")
	if len(parts) != 3 {
		return nil, false
	}
	if parts[0] != "hcs-2" {
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

func BuildTransactionMemo(operation Operation, registryType RegistryType) string {
	operationCode := map[Operation]int{
		OperationRegister: 0,
		OperationUpdate:   1,
		OperationDelete:   2,
		OperationMigrate:  3,
	}

	code := operationCode[operation]
	return fmt.Sprintf("hcs-2:op:%d:%d", code, registryType)
}
