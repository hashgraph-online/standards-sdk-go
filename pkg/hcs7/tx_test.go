package hcs7

import "testing"

func TestBuildSubmitMessageTx(t *testing.T) {
	transaction, err := BuildSubmitMessageTx("0.0.1001", Message{
		P:       "hcs-7",
		Op:      OperationRegister,
		TopicID: "0.0.1002",
		Data: map[string]any{
			"weight": 1,
			"tags":   []string{"registry"},
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

