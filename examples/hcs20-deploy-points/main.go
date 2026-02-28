package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs20"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	if accountID == "" || privateKey == "" {
		fmt.Println("HEDERA_ACCOUNT_ID and HEDERA_PRIVATE_KEY are required")
		return
	}

	client, err := hcs20.NewClient(hcs20.ClientConfig{
		OperatorAccountID:  accountID,
		OperatorPrivateKey: privateKey,
		Network:            "testnet",
	})
	if err != nil {
		panic(err)
	}

	uniqueSuffix := time.Now().UTC().UnixNano()
	tick := fmt.Sprintf("go20%d", uniqueSuffix%1_000_000)

	pointsInfo, err := client.DeployPoints(context.Background(), hcs20.DeployPointsOptions{
		Name:            "Go SDK HCS-20 Example",
		Tick:            tick,
		Max:             "1000000",
		LimitPerMint:    "1000",
		Metadata:        "https://example.com/hcs20-demo",
		UsePrivateTopic: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("deployed HCS-20 points tick=%s topic=%s\n", pointsInfo.Tick, pointsInfo.TopicID)
}
