package hcs27

import (
	"fmt"
	"strconv"
	"strings"
)

func parseCanonicalUint64(fieldName, value string) (uint64, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}
	if value != strings.TrimSpace(value) {
		return 0, fmt.Errorf("%s must be a canonical base-10 string", fieldName)
	}
	if value != "0" && strings.HasPrefix(value, "0") {
		return 0, fmt.Errorf("%s must be a canonical base-10 string", fieldName)
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a canonical base-10 string: %w", fieldName, err)
	}

	return parsed, nil
}

func canonicalUint64(value uint64) string {
	return strconv.FormatUint(value, 10)
}
