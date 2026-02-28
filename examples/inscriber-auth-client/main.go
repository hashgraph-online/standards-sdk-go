package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	if accountID == "" || privateKey == "" {
		fmt.Println("HEDERA_ACCOUNT_ID and HEDERA_PRIVATE_KEY are required")
		return
	}

	authClient := inscriber.NewAuthClient("")
	auth, err := authClient.Authenticate(
		context.Background(),
		accountID,
		privateKey,
		inscriber.NetworkTestnet,
	)
	if err != nil {
		panic(err)
	}

	client, err := inscriber.NewClient(inscriber.Config{
		APIKey:  auth.APIKey,
		Network: inscriber.NetworkTestnet,
	})
	if err != nil {
		panic(err)
	}

	_ = client
	fmt.Println("inscriber client initialized")
}
