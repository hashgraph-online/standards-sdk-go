package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs10"
)

func main() {
	message := hcs10.BuildMessagePayload("0.0.123@0.0.456", "hello from hcs10", "example")
	transaction, err := hcs10.BuildSubmitMessageTx("0.0.123", message, hcs10.BuildTransactionMemo(6, 3))
	if err != nil {
		panic(err)
	}
	fmt.Printf("built hcs10 message tx: %T\n", transaction)
}
