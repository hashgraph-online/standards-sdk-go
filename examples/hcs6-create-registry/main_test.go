package main

import (
	"os"
	"testing"
)

func TestMainCmd(t *testing.T) {
	os.Setenv("HEDERA_ACCOUNT_ID", "0.0.123")
	os.Setenv("HEDERA_PRIVATE_KEY", "302e020100300506032b657004220420d409fcb475960417684d0bfe4c424a13c9e6db58fdff1d5635f11370ded185e9")
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Logf("Recovered from panic in main: %v", recovered)
		}
	}()
	main()
}
