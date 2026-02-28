package hcs2

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/shared"
)

func TestHCS2Integration_EndToEnd(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live Hedera integration tests")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	client, err := NewClient(ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create HCS-2 client: %v", err)
	}

	ctx := context.Background()
	testTTL := int64(3600)

	registryResult, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		RegistryType:        RegistryTypeIndexed,
		TTL:                 testTTL,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}
	t.Logf("created HCS-2 registry topic: %s", registryResult.TopicID)

	targetTopicResult, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		RegistryType: RegistryTypeNonIndexed,
		TTL:          testTTL,
	})
	if err != nil {
		t.Fatalf("failed to create target topic: %v", err)
	}
	targetTopicID := targetTopicResult.TopicID
	t.Logf("created target topic: %s", targetTopicID)

	registerResult, err := client.RegisterEntry(
		ctx,
		registryResult.TopicID,
		RegisterEntryOptions{
			TargetTopicID: targetTopicID,
			Metadata:      "hcs://1/" + targetTopicID,
			Memo:          "go-sdk register",
		},
		"hcs-2",
	)
	if err != nil {
		t.Fatalf("failed to register entry: %v", err)
	}
	t.Logf("registered entry at sequence: %d", registerResult.SequenceNumber)

	time.Sleep(8 * time.Second)

	updateResult, err := client.UpdateEntry(
		ctx,
		registryResult.TopicID,
		UpdateEntryOptions{
			UID:           int64ToString(registerResult.SequenceNumber),
			TargetTopicID: targetTopicID,
			Metadata:      "hcs://1/" + targetTopicID + "@updated",
			Memo:          "go-sdk update",
		},
	)
	if err != nil {
		t.Fatalf("failed to update entry: %v", err)
	}
	t.Logf("updated entry at sequence: %d", updateResult.SequenceNumber)

	deleteResult, err := client.DeleteEntry(
		ctx,
		registryResult.TopicID,
		DeleteEntryOptions{
			UID:  int64ToString(registerResult.SequenceNumber),
			Memo: "go-sdk delete",
		},
	)
	if err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}
	t.Logf("delete operation sequence: %d", deleteResult.SequenceNumber)

	time.Sleep(10 * time.Second)

	registry, err := client.GetRegistry(ctx, registryResult.TopicID, QueryRegistryOptions{
		Order: "asc",
	})
	if err != nil {
		t.Fatalf("failed to query registry: %v", err)
	}

	if registry.RegistryType != RegistryTypeIndexed {
		t.Fatalf("expected indexed registry, got %d", registry.RegistryType)
	}
	if registry.TTL != testTTL {
		t.Fatalf("expected registry TTL %d, got %d", testTTL, registry.TTL)
	}
	if len(registry.Entries) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(registry.Entries))
	}

	foundRegister := false
	foundUpdate := false
	foundDelete := false
	for _, entry := range registry.Entries {
		switch entry.Message.Op {
		case OperationRegister:
			foundRegister = true
		case OperationUpdate:
			foundUpdate = true
		case OperationDelete:
			foundDelete = true
		}
	}

	if !foundRegister || !foundUpdate || !foundDelete {
		t.Fatalf("expected register/update/delete entries; got register=%v update=%v delete=%v", foundRegister, foundUpdate, foundDelete)
	}
}

func int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}
