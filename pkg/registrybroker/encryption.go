package registrybroker

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
)

// RegisterEncryptionKey registers the requested resource.
func (c *RegistryBrokerClient) RegisterEncryptionKey(
	ctx context.Context,
	payload RegisterEncryptionKeyPayload,
) (JSONObject, error) {
	body := bodyMap(payload)
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/encryption/keys",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// EnsureAgentKey performs the requested operation.
func (c *RegistryBrokerClient) EnsureAgentKey(
	ctx context.Context,
	options EnsureAgentKeyOptions,
) (JSONObject, error) {
	if strings.TrimSpace(options.UAID) == "" &&
		strings.TrimSpace(options.LedgerAccountID) == "" &&
		strings.TrimSpace(options.Email) == "" {
		return nil, fmt.Errorf("ensure key requires uaid, ledgerAccountId, or email")
	}

	keyType := strings.TrimSpace(options.KeyType)
	if keyType == "" {
		keyType = "secp256k1"
	}

	publicKey := strings.TrimSpace(options.PublicKey)
	privateKey := strings.TrimSpace(options.PrivateKey)
	if publicKey == "" && privateKey != "" {
		derived, err := derivePublicKeyFromPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		publicKey = derived
	}
	if publicKey == "" && options.GenerateIfMissing {
		generated, err := c.GenerateEncryptionKeyPair()
		if err != nil {
			return nil, err
		}
		publicKey = generated.PublicKey
		privateKey = generated.PrivateKey
	}
	if publicKey == "" {
		return nil, fmt.Errorf("public key material is required")
	}

	registerPayload := RegisterEncryptionKeyPayload{
		KeyType:         keyType,
		PublicKey:       publicKey,
		UAID:            strings.TrimSpace(options.UAID),
		LedgerAccountID: strings.TrimSpace(options.LedgerAccountID),
		LedgerNetwork:   strings.TrimSpace(options.LedgerNetwork),
		Email:           strings.TrimSpace(options.Email),
	}
	result, err := c.RegisterEncryptionKey(ctx, registerPayload)
	if err != nil {
		return nil, err
	}
	if privateKey != "" {
		result["privateKey"] = privateKey
	}
	return result, nil
}

// GenerateEncryptionKeyPair performs the requested operation.
func (c *RegistryBrokerClient) GenerateEncryptionKeyPair() (EphemeralKeyPair, error) {
	return c.GenerateEphemeralKeyPair()
}

// GenerateEphemeralKeyPair performs the requested operation.
func (c *RegistryBrokerClient) GenerateEphemeralKeyPair() (EphemeralKeyPair, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return EphemeralKeyPair{}, err
	}
	return EphemeralKeyPair{
		PrivateKey: hex.EncodeToString(privateKey.Serialize()),
		PublicKey:  hex.EncodeToString(privateKey.PubKey().SerializeCompressed()),
	}, nil
}

// DeriveSharedSecret performs the requested operation.
func (c *RegistryBrokerClient) DeriveSharedSecret(options DeriveSharedSecretOptions) ([]byte, error) {
	privateKeyBytes, err := parseHexString(options.PrivateKey)
	if err != nil {
		return nil, err
	}
	peerPublicKeyBytes, err := parseHexString(options.PeerPublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	peerPublicKey, err := btcec.ParsePubKey(peerPublicKeyBytes)
	if err != nil {
		return nil, err
	}
	sharedSecret := btcec.GenerateSharedSecret(privateKey, peerPublicKey)
	digest := sha256.Sum256(sharedSecret)
	return digest[:], nil
}

// BuildCipherEnvelope builds and returns the configured value.
func (c *RegistryBrokerClient) BuildCipherEnvelope(
	options EncryptCipherEnvelopeOptions,
) (CipherEnvelope, error) {
	if strings.TrimSpace(options.Plaintext) == "" {
		return CipherEnvelope{}, ensureNonEmpty(options.Plaintext, "plaintext")
	}
	if strings.TrimSpace(options.SessionID) == "" {
		return CipherEnvelope{}, ensureNonEmpty(options.SessionID, "sessionId")
	}
	secret := normalizeSharedSecret(options.SharedSecret)

	block, err := aes.NewCipher(secret)
	if err != nil {
		return CipherEnvelope{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return CipherEnvelope{}, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return CipherEnvelope{}, err
	}

	associatedDataSource := strings.TrimSpace(options.AssociatedData)
	if associatedDataSource == "" {
		associatedDataSource = options.SessionID
	}
	associatedData := []byte(associatedDataSource)
	encrypted := gcm.Seal(nil, nonce, []byte(options.Plaintext), associatedData)
	associatedDataEncoded := ""
	if len(associatedData) > 0 {
		associatedDataEncoded = base64.StdEncoding.EncodeToString(associatedData)
	}

	recipients := make([]CipherEnvelopeRecipient, 0, len(options.Recipients))
	for _, recipient := range options.Recipients {
		recipients = append(recipients, CipherEnvelopeRecipient{
			UAID:            recipient.UAID,
			UserID:          recipient.UserID,
			LedgerAccountID: recipient.LedgerAccountID,
			Email:           recipient.Email,
			EncryptedShare:  recipient.EncryptedShare,
		})
	}

	revision := options.Revision
	if revision <= 0 {
		revision = 1
	}

	return CipherEnvelope{
		Algorithm:      "aes-256-gcm",
		Ciphertext:     base64.StdEncoding.EncodeToString(encrypted),
		Nonce:          base64.StdEncoding.EncodeToString(nonce),
		AssociatedData: associatedDataEncoded,
		KeyLocator: JSONObject{
			"sessionId": options.SessionID,
			"revision":  revision,
		},
		Recipients: recipients,
	}, nil
}

// OpenCipherEnvelope performs the requested operation.
func (c *RegistryBrokerClient) OpenCipherEnvelope(
	options DecryptCipherEnvelopeOptions,
) (string, error) {
	secret := normalizeSharedSecret(options.SharedSecret)
	nonce, err := base64.StdEncoding.DecodeString(options.Envelope.Nonce)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(options.Envelope.Ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	var associatedData []byte
	if strings.TrimSpace(options.Envelope.AssociatedData) != "" {
		associatedData, err = base64.StdEncoding.DecodeString(options.Envelope.AssociatedData)
		if err != nil {
			return "", err
		}
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func parseCipherEnvelope(value any) (CipherEnvelope, bool) {
	body, ok := value.(map[string]any)
	if !ok {
		return CipherEnvelope{}, false
	}
	envelope := CipherEnvelope{
		Algorithm:      mapStringField(body, "algorithm"),
		Ciphertext:     mapStringField(body, "ciphertext"),
		Nonce:          mapStringField(body, "nonce"),
		AssociatedData: mapStringField(body, "associatedData"),
	}
	if keyLocator, ok := body["keyLocator"].(map[string]any); ok {
		envelope.KeyLocator = keyLocator
	}
	if recipients, ok := body["recipients"].([]any); ok {
		parsedRecipients := make([]CipherEnvelopeRecipient, 0, len(recipients))
		for _, rawRecipient := range recipients {
			recipientMap, ok := rawRecipient.(map[string]any)
			if !ok {
				continue
			}
			parsedRecipients = append(parsedRecipients, CipherEnvelopeRecipient{
				UAID:            mapStringField(recipientMap, "uaid"),
				UserID:          mapStringField(recipientMap, "userId"),
				LedgerAccountID: mapStringField(recipientMap, "ledgerAccountId"),
				Email:           mapStringField(recipientMap, "email"),
				EncryptedShare:  mapStringField(recipientMap, "encryptedShare"),
			})
		}
		envelope.Recipients = parsedRecipients
	}
	if envelope.Ciphertext == "" || envelope.Nonce == "" {
		return CipherEnvelope{}, false
	}
	return envelope, true
}

func derivePublicKeyFromPrivateKey(privateKeyHex string) (string, error) {
	decoded, err := parseHexString(privateKeyHex)
	if err != nil {
		return "", err
	}
	privateKey, _ := btcec.PrivKeyFromBytes(decoded)
	return hex.EncodeToString(privateKey.PubKey().SerializeCompressed()), nil
}

func parseHexString(value string) ([]byte, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, fmt.Errorf("hex string is required")
	}
	normalized := strings.TrimPrefix(trimmed, "0x")
	decoded, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func normalizeSharedSecret(value []byte) []byte {
	if len(value) == 0 {
		return make([]byte, 32)
	}
	if len(value) == 32 {
		result := make([]byte, 32)
		copy(result, value)
		return result
	}
	digest := sha256.Sum256(value)
	return digest[:]
}

func serialiseAuthConfig(auth *AgentAuthConfig) JSONObject {
	if auth == nil {
		return JSONObject{}
	}
	result := JSONObject{}
	if strings.TrimSpace(auth.Type) != "" {
		result["type"] = strings.TrimSpace(auth.Type)
	}
	if strings.TrimSpace(auth.Token) != "" {
		result["token"] = strings.TrimSpace(auth.Token)
	}
	if strings.TrimSpace(auth.Username) != "" {
		result["username"] = strings.TrimSpace(auth.Username)
	}
	if strings.TrimSpace(auth.Password) != "" {
		result["password"] = strings.TrimSpace(auth.Password)
	}
	if strings.TrimSpace(auth.HeaderName) != "" {
		result["headerName"] = strings.TrimSpace(auth.HeaderName)
	}
	if strings.TrimSpace(auth.HeaderValue) != "" {
		result["headerValue"] = strings.TrimSpace(auth.HeaderValue)
	}
	if len(auth.Headers) > 0 {
		headers := JSONObject{}
		for key, value := range auth.Headers {
			if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
				headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
		if len(headers) > 0 {
			result["headers"] = headers
		}
	}
	return result
}

func mapStringField(source map[string]any, key string) string {
	raw, exists := source[key]
	if !exists || raw == nil {
		return ""
	}
	if typed, ok := raw.(string); ok {
		return typed
	}
	switch typed := raw.(type) {
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}
