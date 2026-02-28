package inscriber

import (
	"strings"
	"testing"
)

func TestNormalizeInscriptionOptionsDefaultsToWebSocket(t *testing.T) {
	options := normalizeInscriptionOptions(
		InscriptionOptions{},
		HederaClientConfig{Network: NetworkTestnet},
	)

	if options.Mode != ModeFile {
		t.Fatalf("expected default mode file, got %s", options.Mode)
	}
	if options.ConnectionMode != ConnectionModeWebSocket {
		t.Fatalf("expected default connection mode websocket, got %s", options.ConnectionMode)
	}
	if options.Network != NetworkTestnet {
		t.Fatalf("expected network to inherit client config, got %s", options.Network)
	}
}

func TestBuildStartInscriptionRequestBuffer(t *testing.T) {
	request, err := buildStartInscriptionRequest(
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   []byte("hello"),
			FileName: "hello.txt",
		},
		"0.0.1234",
		NetworkTestnet,
		InscriptionOptions{
			Mode: ModeFile,
		},
	)
	if err != nil {
		t.Fatalf("buildStartInscriptionRequest failed: %v", err)
	}

	if request.HolderID != "0.0.1234" {
		t.Fatalf("unexpected holderId: %s", request.HolderID)
	}
	if request.File.Type != "base64" {
		t.Fatalf("unexpected file type: %s", request.File.Type)
	}
	if strings.TrimSpace(request.File.Base64) == "" {
		t.Fatalf("expected base64 file content")
	}
}

func TestBuildStartInscriptionRequestHashinalRequiresMetadata(t *testing.T) {
	_, err := buildStartInscriptionRequest(
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   []byte("hello"),
			FileName: "hello.txt",
		},
		"0.0.1234",
		NetworkTestnet,
		InscriptionOptions{
			Mode: ModeHashinal,
		},
	)
	if err == nil {
		t.Fatalf("expected error for missing hashinal metadata")
	}
}

func TestBuildStartInscriptionRequestIncludesMetadataTagsAndChunkSize(t *testing.T) {
	request, err := buildStartInscriptionRequest(
		InscriptionInput{
			Type: InscriptionInputTypeURL,
			URL:  "https://example.com/test.txt",
		},
		"0.0.1234",
		NetworkTestnet,
		InscriptionOptions{
			Mode:      ModeFile,
			ChunkSize: 512,
			Tags:      []string{"alpha", "beta"},
			Metadata: map[string]any{
				"creator":     "0.0.1234",
				"description": "example description",
				"custom":      "value",
			},
		},
	)
	if err != nil {
		t.Fatalf("buildStartInscriptionRequest failed: %v", err)
	}

	if request.ChunkSize != 512 {
		t.Fatalf("expected chunk size 512, got %d", request.ChunkSize)
	}
	if len(request.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(request.Tags))
	}
	if request.Metadata["custom"] != "value" {
		t.Fatalf("expected metadata custom value")
	}
	if request.Creator != "0.0.1234" {
		t.Fatalf("expected creator from metadata, got %q", request.Creator)
	}
	if request.Description != "example description" {
		t.Fatalf("expected description from metadata, got %q", request.Description)
	}
}

func TestBuildBrokerQuoteRequestBulkFiles(t *testing.T) {
	request, err := buildBrokerQuoteRequest(
		InscriptionInput{
			Type: InscriptionInputTypeURL,
			URL:  "https://example.com/archive.zip",
		},
		InscribeViaRegistryBrokerOptions{
			Mode: ModeBulkFiles,
		},
	)
	if err != nil {
		t.Fatalf("buildBrokerQuoteRequest failed: %v", err)
	}

	if request.Mode != ModeBulkFiles {
		t.Fatalf("expected bulk-files mode, got %s", request.Mode)
	}
	if request.InputType != "url" {
		t.Fatalf("expected url input type, got %s", request.InputType)
	}
}

func TestParseJobQuoteUsesTotalCost(t *testing.T) {
	quote, err := parseJobQuote(InscriptionJob{
		TotalCost: 150000000,
	})
	if err != nil {
		t.Fatalf("parseJobQuote failed: %v", err)
	}

	if quote.TotalCostHBAR != "1.5" {
		t.Fatalf("unexpected total cost hbar: %s", quote.TotalCostHBAR)
	}
	if len(quote.Breakdown.Transfers) != 1 {
		t.Fatalf("expected one transfer in breakdown")
	}
}
