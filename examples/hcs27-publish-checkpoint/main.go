package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs27"
)

func main() {
	accountID := os.Getenv("HEDERA_ACCOUNT_ID")
	privateKey := os.Getenv("HEDERA_PRIVATE_KEY")
	checkpointTopicID := os.Getenv("HCS27_CHECKPOINT_TOPIC_ID")
	if accountID == "" || privateKey == "" {
		fmt.Println("HEDERA_ACCOUNT_ID and HEDERA_PRIVATE_KEY are required")
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

	ctx := context.Background()
	if checkpointTopicID == "" {
		topicID, transactionID, createErr := client.CreateCheckpointTopic(ctx, hcs27.CreateTopicOptions{
			TTLSeconds:          3600,
			UseOperatorAsAdmin:  true,
			UseOperatorAsSubmit: true,
		})
		if createErr != nil {
			panic(createErr)
		}
		checkpointTopicID = topicID
		fmt.Printf("created checkpoint topic=%s tx=%s\n", checkpointTopicID, transactionID)
	} else {
		fmt.Printf("using existing checkpoint topic=%s\n", checkpointTopicID)
	}

	inlineMetadata := hcs27.CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: hcs27.StreamID{
			Registry: "ans",
			LogID:    "go-example-inline",
		},
		Log: &hcs27.LogProfile{
			Algorithm: "sha-256",
			Leaf:      "sha256(jcs(event))",
			Merkle:    "rfc9162",
		},
		Root: hcs27.RootCommitment{
			TreeSize:     "1",
			RootHashB64u: hashB64URL("standards-sdk-go-example-inline-root"),
		},
	}

	inlineResult, err := client.PublishCheckpoint(
		ctx,
		checkpointTopicID,
		inlineMetadata,
		"standards-sdk-go inline checkpoint",
		"",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"published inline checkpoint sequence=%d tx=%s\n",
		inlineResult.SequenceNumber,
		inlineResult.TransactionID,
	)

	overflowMetadata := hcs27.CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: hcs27.StreamID{
			Registry: "ans",
			LogID:    "go-example-overflow",
		},
		Log: &hcs27.LogProfile{
			Algorithm: "sha-256",
			Leaf:      strings.Repeat("sha256(jcs(event))-", 90),
			Merkle:    "rfc9162",
		},
		Root: hcs27.RootCommitment{
			TreeSize:     "2",
			RootHashB64u: hashB64URL("standards-sdk-go-example-overflow-root"),
		},
	}

	overflowResult, err := client.PublishCheckpoint(
		ctx,
		checkpointTopicID,
		overflowMetadata,
		"standards-sdk-go overflow checkpoint",
		"",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"published overflow checkpoint sequence=%d tx=%s\n",
		overflowResult.SequenceNumber,
		overflowResult.TransactionID,
	)

	records, err := waitForCheckpoints(ctx, client, checkpointTopicID, 2)
	if err != nil {
		panic(err)
	}

	var overflowReference string
	for _, record := range records {
		var reference string
		if json.Unmarshal(record.Message.Metadata, &reference) == nil && strings.HasPrefix(reference, "hcs://1/") {
			overflowReference = reference
			break
		}
	}
	if overflowReference == "" {
		panic("failed to find HCS-1 metadata reference in fetched checkpoints")
	}

	fmt.Printf("resolved overflow metadata reference=%s\n", overflowReference)
	fmt.Printf("validated checkpoint chain=%t\n", hcs27.ValidateCheckpointChain(records) == nil)
}

func hashB64URL(input string) string {
	sum := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func waitForCheckpoints(
	ctx context.Context,
	client *hcs27.Client,
	topicID string,
	minRecords int,
) ([]hcs27.CheckpointRecord, error) {
	for attempt := 0; attempt < 20; attempt++ {
		records, err := client.GetCheckpoints(ctx, topicID, nil)
		if err == nil && len(records) >= minRecords {
			return records, nil
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("timed out waiting for %d checkpoint records", minRecords)
}
