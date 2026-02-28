package hcs14

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"
)

var fqdnLabelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func normalizeDomain(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, ".")
	return strings.ToLower(trimmed)
}

func isFQDN(value string) bool {
	normalized := normalizeDomain(value)
	if normalized == "" || len(normalized) > 253 || !strings.Contains(normalized, ".") {
		return false
	}

	labels := strings.Split(normalized, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || !fqdnLabelPattern.MatchString(label) {
			return false
		}
	}
	return true
}

func parseSemicolonFields(input string) map[string]string {
	fields := map[string]string{}
	for _, part := range strings.Split(input, ";") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		equalIndex := strings.Index(trimmed, "=")
		if equalIndex <= 0 {
			continue
		}

		key := strings.TrimSpace(trimmed[:equalIndex])
		value := normalizeTXTValue(trimmed[equalIndex+1:])
		if key == "" || value == "" {
			continue
		}
		fields[key] = value
	}

	return fields
}

func normalizeTXTValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 && strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
		return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	return trimmed
}

func nodeDNSTXTLookup(ctx context.Context, hostname string) ([]string, error) {
	records, err := net.DefaultResolver.LookupTXT(ctx, hostname)
	if err != nil {
		var dnsError *net.DNSError
		if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "cannot unmarshal DNS message") {
			return []string{}, nil
		}
		if errors.As(err, &dnsError) {
			if dnsError.IsNotFound {
				return []string{}, nil
			}
		}
		return nil, err
	}
	return records, nil
}
