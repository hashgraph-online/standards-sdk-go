package hcs12

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultTTL int64 = 86400

var topicMemoPattern = regexp.MustCompile(`^hcs-12:1:(\d+):(\d+)$`)

func typeEnum(registryType RegistryType) (int, bool) {
	switch registryType {
	case RegistryTypeAction:
		return 0, true
	case RegistryTypeAssembly:
		return 2, true
	case RegistryTypeHashlinks:
		return 3, true
	default:
		return 0, false
	}
}

// BuildRegistryMemo performs the requested operation.
func BuildRegistryMemo(registryType RegistryType, ttl int64) (string, error) {
	typeValue, ok := typeEnum(registryType)
	if !ok {
		return "", fmt.Errorf("unsupported HCS-12 registry type %q", registryType)
	}
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-12:1:%d:%d", resolvedTTL, typeValue), nil
}

// ParseRegistryMemo performs the requested operation.
func ParseRegistryMemo(memo string) (RegistryType, int64, bool) {
	matches := topicMemoPattern.FindStringSubmatch(strings.TrimSpace(memo))
	if len(matches) != 3 {
		return "", 0, false
	}

	var ttl int64
	_, err := fmt.Sscanf(matches[1], "%d", &ttl)
	if err != nil || ttl <= 0 {
		return "", 0, false
	}
	var enumValue int
	_, err = fmt.Sscanf(matches[2], "%d", &enumValue)
	if err != nil {
		return "", 0, false
	}

	switch enumValue {
	case 0:
		return RegistryTypeAction, ttl, true
	case 2:
		return RegistryTypeAssembly, ttl, true
	case 3:
		return RegistryTypeHashlinks, ttl, true
	default:
		return "", 0, false
	}
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo(operationEnum int, topicTypeEnum int) string {
	return fmt.Sprintf("hcs-12:op:%d:%d", operationEnum, topicTypeEnum)
}

