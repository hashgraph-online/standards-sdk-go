package hcs20

import "testing"

func TestBuildHCS20DeployTx(t *testing.T) {
	transaction, err := BuildHCS20DeployTx(DeployTxParams{
		TopicID: "0.0.12345",
		Name:    "Loyalty",
		Tick:    "loyal",
		Max:     "100000",
		Limit:   "1000",
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if transaction == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS20MintTx(t *testing.T) {
	transaction, err := BuildHCS20MintTx(MintTxParams{
		TopicID: "0.0.12345",
		Tick:    "loyal",
		Amount:  "100",
		To:      "0.0.1001",
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if transaction == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS20TransferTx(t *testing.T) {
	transaction, err := BuildHCS20TransferTx(TransferTxParams{
		TopicID: "0.0.12345",
		Tick:    "loyal",
		Amount:  "10",
		From:    "0.0.1001",
		To:      "0.0.1002",
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if transaction == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS20BurnTx(t *testing.T) {
	transaction, err := BuildHCS20BurnTx(BurnTxParams{
		TopicID: "0.0.12345",
		Tick:    "loyal",
		Amount:  "5",
		From:    "0.0.1001",
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if transaction == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS20RegisterTx(t *testing.T) {
	transaction, err := BuildHCS20RegisterTx(RegisterTxParams{
		RegistryTopicID: "0.0.4362300",
		Name:            "Loyalty Topic",
		Metadata:        "ipfs://meta",
		IsPrivate:       true,
		TopicID:         "0.0.12345",
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if transaction == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS20SubmitMessageTxInvalidPayloadType(t *testing.T) {
	_, err := BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID: "0.0.12345",
		Payload: "invalid-type",
	})
	if err == nil {
		t.Fatal("expected payload type error")
	}
}

func TestBuildHCS20SubmitMessageTxInvalidTopicID(t *testing.T) {
	isPrivate := false
	_, err := BuildHCS20SubmitMessageTx(SubmitMessageTxParams{
		TopicID: "not-a-topic-id",
		Payload: Message{
			Protocol:  "hcs-20",
			Operation: "register",
			Name:      "Bad",
			TopicID:   "0.0.123",
			Private:   &isPrivate,
		},
	})
	if err == nil {
		t.Fatal("expected invalid topic error")
	}
}
