package shared

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const testPrivateKey = "302e020100300506032b65700422042091132178e72057a1d7528025956fe39b0b847f200ab59b2fdd367017f3087137"

var operatorEnvKeys = []string{
	"HEDERA_NETWORK",
	"NETWORK",
	"HEDERA_ACCOUNT_ID",
	"HEDERA_OPERATOR_ID",
	"ACCOUNT_ID",
	"OPERATOR_ID",
	"HEDERA_PRIVATE_KEY",
	"HEDERA_OPERATOR_KEY",
	"PRIVATE_KEY",
	"OPERATOR_KEY",
	"MAINNET_HEDERA_ACCOUNT_ID",
	"MAINNET_HEDERA_OPERATOR_ID",
	"MAINNET_OPERATOR_ID",
	"MAINNET_HEDERA_PRIVATE_KEY",
	"MAINNET_HEDERA_OPERATOR_KEY",
	"MAINNET_OPERATOR_KEY",
	"TESTNET_HEDERA_ACCOUNT_ID",
	"TESTNET_HEDERA_OPERATOR_ID",
	"TESTNET_OPERATOR_ID",
	"TESTNET_HEDERA_PRIVATE_KEY",
	"TESTNET_HEDERA_OPERATOR_KEY",
	"TESTNET_OPERATOR_KEY",
}

func resetOperatorEnv(t *testing.T) {
	t.Helper()
	dotenvLoadOnce = sync.Once{}
	dotenvLoadOnce.Do(func() {})
	for _, key := range operatorEnvKeys {
		t.Setenv(key, "")
	}
}

func TestIsValidEnvKey(t *testing.T) {
	valid := []string{
		"A", "ABC", "a_b", "MY_VAR", "foo_bar", "A1", "A_1_B",
		"HEDERA_NETWORK", "_LEADING_UNDERSCORE",
	}
	for _, key := range valid {
		if !isValidEnvKey(key) {
			t.Fatalf("expected %q to be valid", key)
		}
	}
}

func TestIsValidEnvKeyInvalid(t *testing.T) {
	invalid := []string{
		"", "1ABC", "A B", "A-B", "A.B", "A=B",
	}
	for _, key := range invalid {
		if isValidEnvKey(key) {
			t.Fatalf("expected %q to be invalid", key)
		}
	}
}

func TestFirstNonEmptyEnv(t *testing.T) {
	os.Setenv("_TEST_FIRST_A", "")
	os.Setenv("_TEST_FIRST_B", "hello")
	defer os.Unsetenv("_TEST_FIRST_A")
	defer os.Unsetenv("_TEST_FIRST_B")

	result := firstNonEmptyEnv("_TEST_FIRST_A", "_TEST_FIRST_B")
	if result != "hello" {
		t.Fatalf("expected 'hello', got %q", result)
	}
}

func TestFirstNonEmptyEnvAllEmpty(t *testing.T) {
	result := firstNonEmptyEnv("_TEST_NONEXISTENT_1", "_TEST_NONEXISTENT_2")
	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestFirstNonEmptyEnvTrimsWhitespace(t *testing.T) {
	os.Setenv("_TEST_WS", "   ")
	defer os.Unsetenv("_TEST_WS")

	result := firstNonEmptyEnv("_TEST_WS")
	if result != "" {
		t.Fatalf("expected empty string for whitespace-only, got %q", result)
	}
}

func TestParsePrivateKeyEmpty(t *testing.T) {
	_, err := ParsePrivateKey("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestParsePrivateKeyWhitespace(t *testing.T) {
	_, err := ParsePrivateKey("   ")
	if err == nil {
		t.Fatal("expected error for whitespace key")
	}
}

func TestParsePrivateKeyInvalid(t *testing.T) {
	_, err := ParsePrivateKey("notavalidkey")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestOperatorConfigFromEnvMissingAccountID(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_PRIVATE_KEY", testPrivateKey)

	_, err := OperatorConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for missing account ID")
	}
}

func TestOperatorConfigFromEnvMissingPrivateKey(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_ACCOUNT_ID", "0.0.12345")

	_, err := OperatorConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for missing private key")
	}
}

func TestOperatorConfigFromEnvSuccessTestnet(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_NETWORK", "testnet")
	t.Setenv("HEDERA_ACCOUNT_ID", "0.0.12345")
	t.Setenv("HEDERA_PRIVATE_KEY", testPrivateKey)

	config, err := OperatorConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.AccountID != "0.0.12345" {
		t.Fatalf("expected account ID '0.0.12345', got %q", config.AccountID)
	}
	if config.Network != "testnet" {
		t.Fatalf("expected network 'testnet', got %q", config.Network)
	}
}

func TestOperatorConfigFromEnvScopedMainnet(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_NETWORK", "mainnet")
	t.Setenv("HEDERA_ACCOUNT_ID", "0.0.11111")
	t.Setenv("HEDERA_PRIVATE_KEY", testPrivateKey)
	t.Setenv("MAINNET_HEDERA_ACCOUNT_ID", "0.0.99999")
	t.Setenv("MAINNET_HEDERA_PRIVATE_KEY", testPrivateKey)

	config, err := OperatorConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.AccountID != "0.0.99999" {
		t.Fatalf("expected scoped mainnet account ID '0.0.99999', got %q", config.AccountID)
	}
}

func TestOperatorConfigFromEnvScopedTestnet(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_NETWORK", "testnet")
	t.Setenv("HEDERA_ACCOUNT_ID", "0.0.11111")
	t.Setenv("HEDERA_PRIVATE_KEY", testPrivateKey)
	t.Setenv("TESTNET_HEDERA_ACCOUNT_ID", "0.0.88888")
	t.Setenv("TESTNET_HEDERA_PRIVATE_KEY", testPrivateKey)

	config, err := OperatorConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.AccountID != "0.0.88888" {
		t.Fatalf("expected scoped testnet account ID '0.0.88888', got %q", config.AccountID)
	}
}

func TestOperatorConfigFromEnvDefaultNetwork(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("HEDERA_ACCOUNT_ID", "0.0.12345")
	t.Setenv("HEDERA_PRIVATE_KEY", testPrivateKey)

	config, err := OperatorConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Network != "testnet" {
		t.Fatalf("expected default network 'testnet', got %q", config.Network)
	}
}

func TestOperatorConfigFromEnvFallbackOperatorKeys(t *testing.T) {
	resetOperatorEnv(t)
	t.Setenv("OPERATOR_ID", "0.0.77777")
	t.Setenv("OPERATOR_KEY", testPrivateKey)

	config, err := OperatorConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.AccountID != "0.0.77777" {
		t.Fatalf("expected '0.0.77777', got %q", config.AccountID)
	}
}

func TestLoadDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("_TEST_DOTENV_LOAD=loaded_value\n"), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
	defer os.Unsetenv("_TEST_DOTENV_LOAD")

	result := loadDotEnvFile(envPath)
	if !result {
		t.Fatal("expected loadDotEnvFile to return true")
	}
	if os.Getenv("_TEST_DOTENV_LOAD") != "loaded_value" {
		t.Fatalf("expected 'loaded_value', got %q", os.Getenv("_TEST_DOTENV_LOAD"))
	}
}

func TestLoadDotEnvFileSkipsComments(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env-comments")
	content := "# comment\n\n_TEST_DOTENV_COMMENT=yes\nexport _TEST_DOTENV_EXPORT=exported\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
	defer os.Unsetenv("_TEST_DOTENV_COMMENT")
	defer os.Unsetenv("_TEST_DOTENV_EXPORT")

	result := loadDotEnvFile(envPath)
	if !result {
		t.Fatal("expected loadDotEnvFile to return true")
	}
	if os.Getenv("_TEST_DOTENV_COMMENT") != "yes" {
		t.Fatalf("expected 'yes', got %q", os.Getenv("_TEST_DOTENV_COMMENT"))
	}
	if os.Getenv("_TEST_DOTENV_EXPORT") != "exported" {
		t.Fatalf("expected 'exported', got %q", os.Getenv("_TEST_DOTENV_EXPORT"))
	}
}

func TestLoadDotEnvFileStripsQuotes(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env-quotes")
	content := "_TEST_DOTENV_DQ=\"double-quoted\"\n_TEST_DOTENV_SQ='single-quoted'\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
	defer os.Unsetenv("_TEST_DOTENV_DQ")
	defer os.Unsetenv("_TEST_DOTENV_SQ")

	result := loadDotEnvFile(envPath)
	if !result {
		t.Fatal("expected loadDotEnvFile to return true")
	}
	if os.Getenv("_TEST_DOTENV_DQ") != "double-quoted" {
		t.Fatalf("expected 'double-quoted', got %q", os.Getenv("_TEST_DOTENV_DQ"))
	}
	if os.Getenv("_TEST_DOTENV_SQ") != "single-quoted" {
		t.Fatalf("expected 'single-quoted', got %q", os.Getenv("_TEST_DOTENV_SQ"))
	}
}

func TestLoadDotEnvFileSkipsAlreadySet(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env-skip")
	os.Setenv("_TEST_DOTENV_PREEXIST", "original")
	content := "_TEST_DOTENV_PREEXIST=overridden\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
	defer os.Unsetenv("_TEST_DOTENV_PREEXIST")

	loadDotEnvFile(envPath)
	if os.Getenv("_TEST_DOTENV_PREEXIST") != "original" {
		t.Fatalf("expected 'original' (not overridden), got %q", os.Getenv("_TEST_DOTENV_PREEXIST"))
	}
}

func TestLoadDotEnvFileSkipsInvalidKeys(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env-invalid-keys")
	content := "1BAD=value\n=nokey\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	result := loadDotEnvFile(envPath)
	if result {
		t.Fatal("expected loadDotEnvFile to return false for invalid keys")
	}
	if _, exists := os.LookupEnv("1BAD"); exists {
		t.Fatal("expected invalid key 1BAD to remain unset")
	}
}

func TestLoadDotEnvFileNonexistent(t *testing.T) {
	result := loadDotEnvFile("/tmp/_nonexistent_test_env_file_12345")
	if result {
		t.Fatal("expected loadDotEnvFile to return false for nonexistent file")
	}
}

func TestParsePrivateKeyValidEd25519(t *testing.T) {
	key, err := ParsePrivateKey(testPrivateKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.String() == "" {
		t.Fatal("expected non-empty key string")
	}
}
