package main

import (
	"fmt"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs17"
)

func main() {
	message := hcs17.StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "a3f8ef4f8f7f8c4e0722f2855f3aa85123dc5ef437e7099e06b89cf660dcf3afe855f43de7de6db95f4f8d4ad8198f50",
		Topics:    []string{"0.0.3001", "0.0.3002"},
		AccountID: "0.0.3000",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Memo:      "state hash example",
	}

	tx, err := hcs17.BuildStateHashMessageTx("0.0.3999", message, "hcs-17 example")
	if err != nil {
		panic(err)
	}

	_ = tx
	fmt.Printf("built hcs-17 state hash message for account %s across %d topics\n", message.AccountID, len(message.Topics))
}
