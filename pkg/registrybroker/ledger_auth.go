package registrybroker

import (
	"context"
	"net/http"
	"strings"
)

// CreateLedgerChallenge creates the requested resource.
func (c *RegistryBrokerClient) CreateLedgerChallenge(
	ctx context.Context,
	payload LedgerChallengeRequest,
) (JSONObject, error) {
	if err := ensureNonEmpty(payload.AccountID, "accountId"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(payload.Network, "network"); err != nil {
		return nil, err
	}
	body := JSONObject{
		"accountId": strings.TrimSpace(payload.AccountID),
		"network":   canonicalizeLedgerNetwork(payload.Network),
	}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/auth/ledger/challenge",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// VerifyLedgerChallenge performs the requested operation.
func (c *RegistryBrokerClient) VerifyLedgerChallenge(
	ctx context.Context,
	payload LedgerVerifyRequest,
) (JSONObject, error) {
	if err := ensureNonEmpty(payload.ChallengeID, "challengeId"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(payload.AccountID, "accountId"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(payload.Network, "network"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(payload.Signature, "signature"); err != nil {
		return nil, err
	}

	body := JSONObject{
		"challengeId": strings.TrimSpace(payload.ChallengeID),
		"accountId":   strings.TrimSpace(payload.AccountID),
		"network":     canonicalizeLedgerNetwork(payload.Network),
		"signature":   strings.TrimSpace(payload.Signature),
	}
	if strings.TrimSpace(payload.SignatureKind) != "" {
		body["signatureKind"] = strings.TrimSpace(payload.SignatureKind)
	}
	if strings.TrimSpace(payload.PublicKey) != "" {
		body["publicKey"] = strings.TrimSpace(payload.PublicKey)
	}
	if payload.ExpiresInMinutes != nil {
		body["expiresInMinutes"] = *payload.ExpiresInMinutes
	}

	result, err := c.requestJSON(
		ctx,
		http.MethodPost,
		"/auth/ledger/verify",
		body,
		map[string]string{"content-type": "application/json"},
	)
	if err != nil {
		return nil, err
	}

	if key, ok := result["key"].(string); ok && strings.TrimSpace(key) != "" {
		c.SetLedgerAPIKey(strings.TrimSpace(key))
	}
	return result, nil
}

// AuthenticateWithLedger authenticates the current request.
func (c *RegistryBrokerClient) AuthenticateWithLedger(
	ctx context.Context,
	options LedgerAuthenticationOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(options.AccountID, "accountId"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(options.Network, "network"); err != nil {
		return nil, err
	}
	if options.Sign == nil {
		return nil, ensureNonEmpty("", "sign function")
	}

	challenge, err := c.CreateLedgerChallenge(ctx, LedgerChallengeRequest{
		AccountID: options.AccountID,
		Network:   options.Network,
	})
	if err != nil {
		return nil, err
	}

	message := stringValue(challenge, "message")
	challengeID := stringValue(challenge, "challengeId")
	signResult, err := options.Sign(message)
	if err != nil {
		return nil, err
	}

	return c.VerifyLedgerChallenge(ctx, LedgerVerifyRequest{
		ChallengeID:      challengeID,
		AccountID:        options.AccountID,
		Network:          options.Network,
		Signature:        signResult.Signature,
		SignatureKind:    signResult.SignatureKind,
		PublicKey:        signResult.PublicKey,
		ExpiresInMinutes: options.ExpiresInMinutes,
	})
}

// AuthenticateWithLedgerCredentials authenticates the current request.
func (c *RegistryBrokerClient) AuthenticateWithLedgerCredentials(
	ctx context.Context,
	options LedgerCredentialAuthOptions,
) (JSONObject, error) {
	if options.Sign != nil {
		return c.AuthenticateWithLedger(ctx, LedgerAuthenticationOptions{
			AccountID:        options.AccountID,
			Network:          options.Network,
			ExpiresInMinutes: options.ExpiresInMinutes,
			Sign: func(message string) (LedgerAuthenticationSignerResult, error) {
				return options.Sign(ctx, message)
			},
		})
	}

	if err := ensureNonEmpty(options.Signature, "signature"); err != nil {
		return nil, err
	}

	challenge, err := c.CreateLedgerChallenge(ctx, LedgerChallengeRequest{
		AccountID: options.AccountID,
		Network:   options.Network,
	})
	if err != nil {
		return nil, err
	}

	challengeID := stringValue(challenge, "challengeId")
	result, err := c.VerifyLedgerChallenge(ctx, LedgerVerifyRequest{
		ChallengeID:      challengeID,
		AccountID:        options.AccountID,
		Network:          options.Network,
		Signature:        options.Signature,
		SignatureKind:    options.SignatureKind,
		PublicKey:        options.PublicKey,
		ExpiresInMinutes: options.ExpiresInMinutes,
	})
	if err != nil {
		return nil, err
	}
	if !options.SetAccountHeader {
		return result, nil
	}
	accountID := stringValue(result, "accountId")
	if strings.TrimSpace(accountID) != "" {
		c.SetDefaultHeader("x-account-id", accountID)
	}
	return result, nil
}

func canonicalizeLedgerNetwork(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if normalized == "" {
		return normalized
	}
	switch normalized {
	case "hedera:mainnet", "mainnet":
		return "hedera:mainnet"
	case "hedera:testnet", "testnet":
		return "hedera:testnet"
	default:
		return normalized
	}
}

func stringValue(source JSONObject, key string) string {
	if source == nil {
		return ""
	}
	raw, exists := source[key]
	if !exists || raw == nil {
		return ""
	}
	if typed, ok := raw.(string); ok {
		return typed
	}
	return ""
}
