package registrybroker

import (
	"context"
	"fmt"
	"strings"
)

func (h *ChatConversationHandle) Send(
	ctx context.Context,
	plaintext string,
	streaming *bool,
	auth *AgentAuthConfig,
) (JSONObject, error) {
	if h == nil || h.client == nil {
		return nil, ensureNonEmpty("", "conversation handle client")
	}
	requestAuth := auth
	if requestAuth == nil {
		requestAuth = h.defaultAuth
	}

	payload := SendMessageRequestPayload{
		Message:   plaintext,
		SessionID: h.SessionID,
		UAID:      h.uaid,
		AgentURL:  h.agentURL,
		Streaming: streaming,
		Auth:      requestAuth,
	}
	if strings.EqualFold(strings.TrimSpace(h.Mode), "encrypted") && len(h.sharedSecret) > 0 {
		recipients := cloneRecipients(h.recipients)
		if len(recipients) == 0 && h.identity != nil {
			recipients = []CipherEnvelopeRecipient{{
				UAID:            h.identity.UAID,
				LedgerAccountID: h.identity.LedgerAccountID,
				UserID:          h.identity.UserID,
				Email:           h.identity.Email,
			}}
		}
		if len(recipients) == 0 {
			return nil, fmt.Errorf("recipients are required for encrypted chat payloads")
		}
		payload.Message = "[ciphertext omitted]"
		payload.Encryption = &SendMessageEncryptionOptions{
			Plaintext:    plaintext,
			SessionID:    h.SessionID,
			SharedSecret: cloneBytes(h.sharedSecret),
			Recipients:   recipients,
		}
	}

	return h.client.SendMessage(ctx, payload)
}

func (h *ChatConversationHandle) FetchHistory(
	ctx context.Context,
	options ChatHistoryFetchOptions,
) (JSONObject, error) {
	if h == nil || h.client == nil {
		return nil, ensureNonEmpty("", "conversation handle client")
	}
	requestOptions := options
	if strings.EqualFold(strings.TrimSpace(h.Mode), "encrypted") && len(h.sharedSecret) > 0 {
		if requestOptions.Decrypt == nil {
			decrypt := true
			requestOptions.Decrypt = &decrypt
		}
		if len(requestOptions.SharedSecret) == 0 {
			requestOptions.SharedSecret = cloneBytes(h.sharedSecret)
		}
		if requestOptions.Identity == nil {
			requestOptions.Identity = cloneRecipientIdentity(h.identity)
		}
	}
	return h.client.FetchHistorySnapshot(ctx, h.SessionID, requestOptions)
}

func addObjectString(target JSONObject, key string, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		target[key] = trimmed
	}
}

func identitiesMatch(a *RecipientIdentity, b *RecipientIdentity) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if strings.EqualFold(a.UAID, b.UAID) && strings.TrimSpace(a.UAID) != "" {
		return true
	}
	if strings.EqualFold(a.LedgerAccountID, b.LedgerAccountID) && strings.TrimSpace(a.LedgerAccountID) != "" {
		return true
	}
	if strings.TrimSpace(a.UserID) != "" && strings.TrimSpace(a.UserID) == strings.TrimSpace(b.UserID) {
		return true
	}
	if strings.EqualFold(a.Email, b.Email) && strings.TrimSpace(a.Email) != "" {
		return true
	}
	return false
}

func cloneRecipientIdentity(value *RecipientIdentity) *RecipientIdentity {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	result := make([]byte, len(value))
	copy(result, value)
	return result
}
