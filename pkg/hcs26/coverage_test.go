package hcs26

import (
"context"
"testing"
)

func TestCovNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil { t.Fatal("expected err") }

	client, err := NewClient(ClientConfig{Network: "testnet"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if client.MirrorClient() == nil { t.Fatal("expected mirror") }
}

func TestCovOperationsFailure(t *testing.T) {
	client, _ := NewClient(ClientConfig{Network: "testnet"})
	ctx := context.Background()

	_, err := client.ResolveDiscoveryRecord(ctx, "invalid", 1, 10)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.ListVersionRegisters(ctx, "invalid", 1, 10)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.GetLatestVersionRegister(ctx, "invalid", int64(1))
	if err == nil { t.Fatal("expected fail") }

	_, _, err = client.ResolveManifest(ctx, "invalid")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.ResolveSkill(ctx, "invalid", 1, 10)
	if err == nil { t.Fatal("expected fail") }
}

func TestCovHelpers(t *testing.T) {
	v, ok := toInt64(float64(42))
	if !ok || v != 42 { t.Fatal("expected 42") }

	v2, ok2 := toInt64(int64(100))
	if !ok2 || v2 != 100 { t.Fatal("expected 100") }

	_, ok3 := toInt64("not a number")
	if ok3 { t.Fatal("expected false") }

	s := readString(map[string]any{"key": "val"}, "key")
	if s != "val" { t.Fatal("expected val") }

	s2 := readString(map[string]any{}, "missing")
	if s2 != "" { t.Fatal("expected empty") }

	seq := sequenceFromPayload(map[string]any{"sequence_number": float64(5)}, 0)
	if seq != 5 { t.Fatal("expected 5") }

	seq2 := sequenceFromPayload(map[string]any{}, 10)
	if seq2 != 10 { t.Fatal("expected 10") }
}

func TestCovValidation(t *testing.T) {
	err := validateDiscoveryMetadata(map[string]any{})
	if err == nil { t.Fatal("expected error") }

	err = validateManifest(SkillManifest{})
	if err == nil { t.Fatal("expected error") }
}

func TestCovDecodePayload(t *testing.T) {
	_, err := decodePayload("not-base64!!!")
	if err == nil { t.Fatal("expected err") }
}
