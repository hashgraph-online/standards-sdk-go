package hcs14

import (
	"regexp"
	"strings"
)

var (
	hederaCAIP10Pattern = regexp.MustCompile(`^hedera:(mainnet|testnet|previewnet|devnet):\d+\.\d+\.\d+$`)
	eip155CAIP10Pattern = regexp.MustCompile(`^eip155:\d+:0x[0-9a-fA-F]{40}$`)
)

func isHederaCAIP10(value string) bool {
	return hederaCAIP10Pattern.MatchString(strings.TrimSpace(value))
}

func isEIP155CAIP10(value string) bool {
	return eip155CAIP10Pattern.MatchString(strings.TrimSpace(value))
}
