package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs27"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	checkpointTopicID := os.Getenv("HCS27_CHECKPOINT_TOPIC_ID")
	if accountID == "" || privateKey == "" || checkpointTopicID == "" {
		fmt.Println("HEDERA_ACCOUNT_ID, HEDERA_PRIVATE_KEY, and HCS27_CHECKPOINT_TOPIC_ID are required")
		return
	}

	client, err := hcs27.NewClient(hcs27.ClientConfig{
		OperatorAccountID:  accountID,
		OperatorPrivateKey: privateKey,
		Network:            "testnet",
	})
	if err != nil {
		panic(err)
	}

	root := hashB64URL("standards-sdk-go-example-root")
	metadata := hcs27.CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: hcs27.StreamID{
			Registry: "ans",
			LogID:    "default",
		},
		Log: &hcs27.LogProfile{
			Algorithm: "sha-256",
			Leaf:      "sha256(jcs(event))",
			Merkle:    "rfc6962",
		},
		Root: hcs27.RootCommitment{
			TreeSize:    1,
			RootHashB64: root,
		},
		BatchRange: hcs27.BatchRange{
			Start: 1,
			End:   1,
		},
	}

	result, err := client.PublishCheckpoint(
		context.Background(),
		checkpointTopicID,
		metadata,
		"standards-sdk-go checkpoint",
		"",
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("published checkpoint sequence=%d tx=%s\n", result.SequenceNumber, result.TransactionID)
}

func hashB64URL(input string) string {
	sum := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
