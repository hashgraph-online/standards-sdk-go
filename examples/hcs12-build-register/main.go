package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs12"
)

func main() {
	payload := map[string]any{
		"p":       "hcs-12",
		"op":      "register",
		"name":    "demo-action",
		"version": "1.0.0",
	}
	transaction, err := hcs12.BuildSubmitMessageTx("0.0.123", payload, hcs12.BuildTransactionMemo(0, 0))
	if err != nil {
		panic(err)
	}
	fmt.Printf("built hcs12 register tx: %T\n", transaction)
}
