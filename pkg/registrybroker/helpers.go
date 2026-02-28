package registrybroker

import (
	"fmt"
	"strings"
)

func ensureNonEmpty(value string, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
