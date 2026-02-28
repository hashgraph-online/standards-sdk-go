package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs2"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	if accountID == "" || privateKey == "" {
		fmt.Println("HEDERA_ACCOUNT_ID and HEDERA_PRIVATE_KEY are required")
		return
	}

	client, err := hcs2.NewClient(hcs2.ClientConfig{
		OperatorAccountID:  accountID,
		OperatorPrivateKey: privateKey,
		Network:            "testnet",
	})
	if err != nil {
		panic(err)
	}

	result, err := client.CreateRegistry(context.Background(), hcs2.CreateRegistryOptions{
		RegistryType:        hcs2.RegistryTypeIndexed,
		TTL:                 86400,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("created hcs2 registry topic %s (%s)\n", result.TopicID, result.TransactionID)
}
