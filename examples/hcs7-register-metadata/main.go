package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs7"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	if accountID == "" || privateKey == "" {
		fmt.Println("HEDERA_ACCOUNT_ID and HEDERA_PRIVATE_KEY are required")
		return
	}

	client, err := hcs7.NewClient(hcs7.ClientConfig{
		OperatorAccountID:  accountID,
		OperatorPrivateKey: privateKey,
		Network:            "testnet",
	})
	if err != nil {
		panic(err)
	}

	registry, err := client.CreateRegistry(context.Background(), hcs7.CreateRegistryOptions{
		TTL:                 86400,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		panic(err)
	}

	target, err := client.CreateRegistry(context.Background(), hcs7.CreateRegistryOptions{
		TTL:                 86400,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		panic(err)
	}

	result, err := client.RegisterMetadata(context.Background(), hcs7.RegisterMetadataOptions{
		RegistryTopicID: registry.TopicID,
		MetadataTopicID: target.TopicID,
		Weight:          100,
		Tags:            []string{"resolver", "evm"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("registered hcs7 metadata seq=%d tx=%s\n", result.SequenceNumber, result.TransactionID)
}
