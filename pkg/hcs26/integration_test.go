package hcs26

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs2"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestHCS26Integration_ResolveSkill(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	hcs2Client, err := hcs2.NewClient(hcs2.ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create hcs2 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	directoryResult, err := hcs2Client.CreateRegistry(ctx, hcs2.CreateRegistryOptions{
		RegistryType:        hcs2.RegistryTypeIndexed,
		TTL:                 86400,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create directory topic: %v", err)
	}
	versionResult, err := hcs2Client.CreateRegistry(ctx, hcs2.CreateRegistryOptions{
		RegistryType:        hcs2.RegistryTypeNonIndexed,
		TTL:                 86400,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create version registry topic: %v", err)
	}

	hederaClient, err := shared.NewHederaClient(operatorConfig.Network)
	if err != nil {
		t.Fatalf("failed to create hedera client: %v", err)
	}
	operatorID, err := hedera.AccountIDFromString(operatorConfig.AccountID)
	if err != nil {
		t.Fatalf("invalid operator id: %v", err)
	}
	operatorKey, err := shared.ParsePrivateKey(operatorConfig.PrivateKey)
	if err != nil {
		t.Fatalf("invalid operator key: %v", err)
	}
	hederaClient.SetOperator(operatorID, operatorKey)

	manifestTopicTx := hedera.NewTopicCreateTransaction()
	manifestTopicResp, err := manifestTopicTx.Execute(hederaClient)
	if err != nil {
		t.Fatalf("failed to create manifest topic: %v", err)
	}
	manifestTopicReceipt, err := manifestTopicResp.GetReceipt(hederaClient)
	if err != nil {
		t.Fatalf("failed to fetch manifest topic receipt: %v", err)
	}
	if manifestTopicReceipt.TopicID == nil {
		t.Fatalf("manifest topic ID missing")
	}
	manifestTopicID := manifestTopicReceipt.TopicID.String()

	manifestPayload := SkillManifest{
		Name:        "go-hcs26-skill",
		Description: "integration skill",
		Version:     "1.0.0",
		License:     "Apache-2.0",
		Author:      "go-sdk",
		Files: []ManifestFile{
			{
				Path:   "SKILL.md",
				HRL:    "hcs://1/" + manifestTopicID,
				SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Mime:   "text/markdown",
			},
		},
	}
	manifestBytes, err := json.Marshal(manifestPayload)
	if err != nil {
		t.Fatalf("failed to encode manifest payload: %v", err)
	}
	manifestSubmit := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(*manifestTopicReceipt.TopicID).
		SetMessage(manifestBytes)
	manifestResp, err := manifestSubmit.Execute(hederaClient)
	if err != nil {
		t.Fatalf("failed to submit manifest payload: %v", err)
	}
	if _, err := manifestResp.GetReceipt(hederaClient); err != nil {
		t.Fatalf("failed to get manifest submit receipt: %v", err)
	}

	directoryTopicID, err := hedera.TopicIDFromString(directoryResult.TopicID)
	if err != nil {
		t.Fatalf("invalid directory topic id: %v", err)
	}
	discoveryRegisterPayload := map[string]any{
		"p":          "hcs-26",
		"op":         "register",
		"t_id":       versionResult.TopicID,
		"account_id": operatorConfig.AccountID,
		"metadata": map[string]any{
			"name":        "go-hcs26-skill",
			"description": "integration skill",
			"author":      "go-sdk",
			"license":     "Apache-2.0",
		},
	}
	discoveryBytes, err := json.Marshal(discoveryRegisterPayload)
	if err != nil {
		t.Fatalf("failed to encode discovery register payload: %v", err)
	}
	discoveryResp, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(directoryTopicID).
		SetMessage(discoveryBytes).
		Execute(hederaClient)
	if err != nil {
		t.Fatalf("failed to submit discovery register payload: %v", err)
	}
	discoveryReceipt, err := discoveryResp.GetReceipt(hederaClient)
	if err != nil {
		t.Fatalf("failed to get discovery register receipt: %v", err)
	}
	skillUID := int64(discoveryReceipt.TopicSequenceNumber)
	if skillUID <= 0 {
		t.Fatalf("unexpected skill uid sequence: %d", skillUID)
	}

	versionTopicID, err := hedera.TopicIDFromString(versionResult.TopicID)
	if err != nil {
		t.Fatalf("invalid version topic id: %v", err)
	}
	versionRegisterPayload := map[string]any{
		"p":         "hcs-26",
		"op":        "register",
		"skill_uid": skillUID,
		"version":   "1.0.0",
		"t_id":      manifestTopicID,
		"status":    "active",
	}
	versionBytes, err := json.Marshal(versionRegisterPayload)
	if err != nil {
		t.Fatalf("failed to encode version register payload: %v", err)
	}
	versionResp, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(versionTopicID).
		SetMessage(versionBytes).
		Execute(hederaClient)
	if err != nil {
		t.Fatalf("failed to submit version register payload: %v", err)
	}
	if _, err := versionResp.GetReceipt(hederaClient); err != nil {
		t.Fatalf("failed to get version register receipt: %v", err)
	}

	hcs26Client, err := NewClient(ClientConfig{
		Network: operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create hcs26 client: %v", err)
	}

	var resolved *ResolvedSkill
	for attempt := 0; attempt < 20; attempt++ {
		resolved, err = hcs26Client.ResolveSkill(ctx, directoryResult.TopicID, skillUID, 500)
		if err == nil && resolved != nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to resolve skill: %v", err)
	}
	if resolved == nil {
		t.Fatalf("expected resolved skill")
	}
	if resolved.Discovery.VersionRegistry != versionResult.TopicID {
		t.Fatalf("unexpected version registry topic: got %s want %s", resolved.Discovery.VersionRegistry, versionResult.TopicID)
	}
}
