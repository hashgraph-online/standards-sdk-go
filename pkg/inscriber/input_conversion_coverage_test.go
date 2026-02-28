package inscriber

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildStartInscriptionRequest(t *testing.T) {
	inputStr := "test data"
	req, err := buildStartInscriptionRequest(InscriptionInput{
		Type: InscriptionInputTypeBuffer,
		Buffer: []byte(inputStr),
		MimeType: "text/plain",
		FileName: "test.txt",
	}, "0.0.123", NetworkTestnet, InscriptionOptions{})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.File.Base64 != base64.StdEncoding.EncodeToString([]byte(inputStr)) {
		t.Fatal("unexpected data")
	}
	if req.File.MimeType != "text/plain" {
		t.Fatal("unexpected mime")
	}

	reqEmpty, err := buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeURL, URL: "http://"}, "0.0.123", NetworkTestnet, InscriptionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqEmpty.Mode != ModeFile {
		t.Fatal("expected default mode")
	}
}

func TestConvertFilePathToBase64(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("hello"), 0644)

	base64Data, filename, mime, err := convertFilePathToBase64(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base64Data != base64.StdEncoding.EncodeToString([]byte("hello")) {
		t.Fatal("unexpected data")
	}
	if filename != "test.txt" {
		t.Fatal("unexpected filename")
	}
	if mime != "text/plain" {
		t.Fatal("unexpected mime")
	}

	_, _, _, err = convertFilePathToBase64("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEncodeBufferToBase64(t *testing.T) {
	str := encodeBufferToBase64([]byte("hello"))
	if str != base64.StdEncoding.EncodeToString([]byte("hello")) {
		t.Fatal("unexpected")
	}
}

func TestGuessMimeTypeFromName(t *testing.T) {
	if guessMimeTypeFromName("test.jpg") != "image/jpeg" {
		t.Fatal("unexpected")
	}
	if guessMimeTypeFromName("test.png") != "image/png" {
		t.Fatal("unexpected")
	}
	if guessMimeTypeFromName("test.json") != "application/json" {
		t.Fatal("unexpected")
	}
	if guessMimeTypeFromName("test.txt") != "text/plain" {
		t.Fatal("unexpected")
	}
	if guessMimeTypeFromName("test.pdf") != "application/pdf" {
		t.Fatal("unexpected")
	}
	if guessMimeTypeFromName("test") != "application/octet-stream" {
		t.Fatal("unexpected")
	}
}

func TestStringOrDefault(t *testing.T) {
	m := map[string]any{"key": "val"}
	if stringOrDefault(m, "key", "def") != "val" {
		t.Fatal("unexpected")
	}
	if stringOrDefault(m, "missing", "def") != "def" {
		t.Fatal("unexpected")
	}
	if stringOrDefault(nil, "key", "def") != "def" {
		t.Fatal("unexpected")
	}
}

func TestBuildStartInscriptionRequestHashinal(t *testing.T) {
	_, err := buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, "0.0.1", NetworkTestnet, InscriptionOptions{Mode: ModeHashinal})
	if err == nil {
		t.Fatal("expected error missing metadataObject")
	}

	_, err = buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, "0.0.1", NetworkTestnet, InscriptionOptions{
		Mode: ModeHashinal,
		Metadata: map[string]any{"name": "test"},
	})
	if err == nil {
		t.Fatal("expected error missing other required metadata")
	}

	req, err := buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, "0.0.1", NetworkTestnet, InscriptionOptions{
		Mode: ModeHashinal,
		Metadata: map[string]any{
			"name": "test",
			"creator": "me",
			"description": "desc",
			"type": "hashinal",
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if req.Creator != "me" || req.Description != "desc" {
		t.Fatal("expected creator and description to be set")
	}

	// Test missing holderId
	_, err = buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeFile}, "", NetworkTestnet, InscriptionOptions{})
	if err == nil {
		t.Fatal("expected error on empty holder id")
	}

	// Test error on url without url field
	_, err = buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeURL}, "0.0.1", NetworkTestnet, InscriptionOptions{})
	if err == nil {
		t.Fatal("expected err")
	}

	// Test error on buffer without buffer field
	_, err = buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputTypeBuffer}, "0.0.1", NetworkTestnet, InscriptionOptions{})
	if err == nil {
		t.Fatal("expected err")
	}

	// Test unsupported type
	_, err = buildStartInscriptionRequest(InscriptionInput{Type: InscriptionInputType("unsupported")}, "0.0.1", NetworkTestnet, InscriptionOptions{})
	if err == nil {
		t.Fatal("expected err on unsupported type")
	}
}
