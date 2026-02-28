package inscriber

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/shared"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	normalizedBaseURL := strings.TrimSpace(baseURL)
	if normalizedBaseURL == "" {
		normalizedBaseURL = "https://kiloscribe.com"
	}
	normalizedBaseURL = strings.TrimRight(normalizedBaseURL, "/")
	if strings.HasSuffix(normalizedBaseURL, "/api") {
		normalizedBaseURL = strings.TrimSuffix(normalizedBaseURL, "/api")
	}

	return &AuthClient{
		baseURL: normalizedBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AuthClient) Authenticate(
	ctx context.Context,
	accountID string,
	privateKey string,
	network Network,
) (AuthResult, error) {
	requestSignatureURL := c.baseURL + "/api/auth/request-signature"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestSignatureURL, nil)
	if err != nil {
		return AuthResult{}, err
	}
	request.Header.Set("x-session", accountID)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to request signature challenge: %w", err)
	}
	defer response.Body.Close()

	challengeBody, err := io.ReadAll(response.Body)
	if err != nil {
		return AuthResult{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return AuthResult{}, fmt.Errorf(
			"request-signature failed with status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(challengeBody)),
		)
	}

	var challenge struct {
		Message json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(challengeBody, &challenge); err != nil {
		return AuthResult{}, fmt.Errorf("failed to decode signature challenge: %w", err)
	}
	if len(challenge.Message) == 0 {
		return AuthResult{}, fmt.Errorf("signature challenge did not include message")
	}

	signingPayload, authDataValue, err := normalizeChallengeMessage(challenge.Message)
	if err != nil {
		return AuthResult{}, err
	}

	key, err := shared.ParsePrivateKey(privateKey)
	if err != nil {
		return AuthResult{}, err
	}

	signatureBytes := key.Sign([]byte(signingPayload))

	authPayload := map[string]any{
		"authData": map[string]any{
			"id":        accountID,
			"signature": hex.EncodeToString(signatureBytes),
			"data":      authDataValue,
			"network":   string(network),
		},
		"include": "apiKey",
	}

	payloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return AuthResult{}, err
	}

	authURL := c.baseURL + "/api/auth/authenticate"
	authRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		authURL,
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return AuthResult{}, err
	}
	authRequest.Header.Set("Content-Type", "application/json")

	authResponse, err := c.httpClient.Do(authRequest)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to authenticate inscription API client: %w", err)
	}
	defer authResponse.Body.Close()

	authBody, err := io.ReadAll(authResponse.Body)
	if err != nil {
		return AuthResult{}, err
	}
	if authResponse.StatusCode < 200 || authResponse.StatusCode >= 300 {
		return AuthResult{}, fmt.Errorf(
			"authenticate failed with status %d: %s",
			authResponse.StatusCode,
			strings.TrimSpace(string(authBody)),
		)
	}

	var result struct {
		APIKey string `json:"apiKey"`
		User   struct {
			SessionToken string `json:"sessionToken"`
		} `json:"user"`
	}
	if err := json.Unmarshal(authBody, &result); err != nil {
		return AuthResult{}, fmt.Errorf("failed to decode authenticate response: %w", err)
	}
	if strings.TrimSpace(result.User.SessionToken) == "" {
		return AuthResult{}, fmt.Errorf("authenticate response did not include session token")
	}
	if strings.TrimSpace(result.APIKey) == "" {
		return AuthResult{}, fmt.Errorf("authenticate response did not include api key")
	}

	return AuthResult{APIKey: result.APIKey}, nil
}

func normalizeChallengeMessage(raw json.RawMessage) (string, any, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return "", nil, fmt.Errorf("signature challenge message cannot be empty")
	}

	if strings.HasPrefix(trimmed, "\"") {
		var challengeString string
		if err := json.Unmarshal(raw, &challengeString); err != nil {
			return "", nil, fmt.Errorf("failed to decode string challenge: %w", err)
		}
		if strings.TrimSpace(challengeString) == "" {
			return "", nil, fmt.Errorf("signature challenge string cannot be empty")
		}
		return challengeString, challengeString, nil
	}

	var challengeObject any
	if err := json.Unmarshal(raw, &challengeObject); err != nil {
		return "", nil, fmt.Errorf("failed to decode object challenge: %w", err)
	}

	normalizedBytes, err := json.Marshal(challengeObject)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode object challenge: %w", err)
	}

	return string(normalizedBytes), challengeObject, nil
}
