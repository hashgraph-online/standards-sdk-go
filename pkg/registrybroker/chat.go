package registrybroker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (c *RegistryBrokerClient) FetchHistorySnapshot(
	ctx context.Context,
	sessionID string,
	options ChatHistoryFetchOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(sessionID, "sessionId"); err != nil {
		return nil, err
	}
	path := "/chat/session/" + percentPath(sessionID) + "/history"
	snapshot, err := c.requestJSON(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return c.AttachDecryptedHistory(sessionID, snapshot, options), nil
}

func (c *RegistryBrokerClient) AttachDecryptedHistory(
	sessionID string,
	snapshot JSONObject,
	options ChatHistoryFetchOptions,
) JSONObject {
	result := JSONObject{}
	for key, value := range snapshot {
		result[key] = value
	}

	shouldDecrypt := false
	if options.Decrypt != nil {
		shouldDecrypt = *options.Decrypt
	}
	if !shouldDecrypt {
		return result
	}

	historyRaw, ok := snapshot["history"].([]any)
	if !ok || len(historyRaw) == 0 {
		result["decryptedHistory"] = []any{}
		return result
	}

	context := c.ResolveDecryptionContext(sessionID, options)
	decrypted := make([]any, 0, len(historyRaw))
	for _, raw := range historyRaw {
		entryMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		entry := JSONObject(entryMap)
		plaintext := c.DecryptHistoryEntryFromContext(entry, context)
		decrypted = append(decrypted, JSONObject{
			"entry":     entry,
			"plaintext": plaintext,
		})
	}
	result["decryptedHistory"] = decrypted
	return result
}

func (c *RegistryBrokerClient) RegisterConversationContextForEncryption(
	context ConversationContextInput,
) {
	normalized := ConversationContextState{
		SessionID:    context.SessionID,
		SharedSecret: cloneBytes(context.SharedSecret),
		Identity:     cloneRecipientIdentity(context.Identity),
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	existing := c.conversationContexts[context.SessionID]
	replaced := false
	for index, current := range existing {
		if identitiesMatch(current.Identity, normalized.Identity) {
			existing[index] = normalized
			replaced = true
			break
		}
	}
	if !replaced {
		existing = append(existing, normalized)
	}
	c.conversationContexts[context.SessionID] = existing
}

func (c *RegistryBrokerClient) ResolveDecryptionContext(
	sessionID string,
	options ChatHistoryFetchOptions,
) *ConversationContextState {
	if len(options.SharedSecret) > 0 {
		return &ConversationContextState{
			SessionID:    sessionID,
			SharedSecret: cloneBytes(options.SharedSecret),
			Identity:     cloneRecipientIdentity(options.Identity),
		}
	}
	c.mutex.RLock()
	contexts := c.conversationContexts[sessionID]
	c.mutex.RUnlock()
	if len(contexts) == 0 {
		return nil
	}
	if options.Identity != nil {
		for _, current := range contexts {
			if identitiesMatch(current.Identity, options.Identity) {
				cloned := current
				cloned.SharedSecret = cloneBytes(current.SharedSecret)
				return &cloned
			}
		}
	}
	cloned := contexts[0]
	cloned.SharedSecret = cloneBytes(contexts[0].SharedSecret)
	return &cloned
}

func (c *RegistryBrokerClient) DecryptHistoryEntryFromContext(
	entry JSONObject,
	context *ConversationContextState,
) *string {
	if entry == nil {
		return nil
	}
	if context == nil {
		content := stringField(entry, "content")
		if strings.TrimSpace(content) == "" {
			return nil
		}
		return &content
	}

	cipherEnvelopeRaw, hasCipher := entry["cipherEnvelope"]
	if !hasCipher || cipherEnvelopeRaw == nil {
		content := stringField(entry, "content")
		if strings.TrimSpace(content) == "" {
			return nil
		}
		return &content
	}

	cipherEnvelope, ok := parseCipherEnvelope(cipherEnvelopeRaw)
	if !ok {
		return nil
	}

	plaintext, err := c.OpenCipherEnvelope(DecryptCipherEnvelopeOptions{
		Envelope:     cipherEnvelope,
		SharedSecret: context.SharedSecret,
	})
	if err != nil {
		return nil
	}
	return &plaintext
}

func (c *RegistryBrokerClient) CreateSession(
	ctx context.Context,
	payload CreateSessionRequestPayload,
	allowHistoryAutoTopUp bool,
) (JSONObject, error) {
	body := JSONObject{}
	addObjectString(body, "uaid", payload.UAID)
	addObjectString(body, "agentUrl", payload.AgentURL)
	addObjectString(body, "senderUaid", payload.SenderUAID)
	if payload.HistoryTTLSeconds != nil {
		body["historyTtlSeconds"] = *payload.HistoryTTLSeconds
	}
	if payload.EncryptionRequested != nil {
		body["encryptionRequested"] = *payload.EncryptionRequested
	}
	if payload.Auth != nil {
		body["auth"] = serialiseAuthConfig(payload.Auth)
	}

	result, err := c.requestJSON(
		ctx,
		http.MethodPost,
		"/chat/session",
		body,
		map[string]string{"content-type": "application/json"},
	)
	if err == nil {
		return result, nil
	}
	if !allowHistoryAutoTopUp || !c.shouldAutoTopUpHistory(payload, err) {
		return nil, err
	}
	if topUpErr := c.executeHistoryAutoTopUp(ctx, "chat.session"); topUpErr != nil {
		return nil, topUpErr
	}
	return c.CreateSession(ctx, payload, false)
}

func (c *RegistryBrokerClient) StartChat(
	ctx context.Context,
	options StartChatOptions,
) (*ChatConversationHandle, error) {
	if strings.TrimSpace(options.UAID) != "" {
		return c.StartConversation(ctx, StartConversationOptions{
			UAID:              options.UAID,
			SenderUAID:        options.SenderUAID,
			HistoryTTLSeconds: options.HistoryTTLSeconds,
			Auth:              options.Auth,
			Encryption:        options.Encryption,
			OnSessionCreated:  options.OnSessionCreated,
		})
	}
	if strings.TrimSpace(options.AgentURL) == "" {
		return nil, fmt.Errorf("startChat requires either uaid or agentUrl")
	}

	payload := CreateSessionRequestPayload{
		AgentURL:          options.AgentURL,
		SenderUAID:        options.SenderUAID,
		HistoryTTLSeconds: options.HistoryTTLSeconds,
		Auth:              options.Auth,
	}
	session, err := c.CreateSession(ctx, payload, true)
	if err != nil {
		return nil, err
	}
	if options.OnSessionCreated != nil {
		options.OnSessionCreated(stringField(session, "sessionId"))
	}
	return c.CreatePlaintextConversationHandle(
		stringField(session, "sessionId"),
		encryptionSummaryFromResponse(session),
		options.Auth,
		options.UAID,
		options.AgentURL,
	), nil
}

func (c *RegistryBrokerClient) StartConversation(
	ctx context.Context,
	options StartConversationOptions,
) (*ChatConversationHandle, error) {
	preference := conversationPreference(options.Encryption)
	if preference == "disabled" {
		requestEncryption := false
		session, err := c.CreateSession(ctx, CreateSessionRequestPayload{
			UAID:                options.UAID,
			SenderUAID:          options.SenderUAID,
			HistoryTTLSeconds:   options.HistoryTTLSeconds,
			Auth:                options.Auth,
			EncryptionRequested: &requestEncryption,
		}, true)
		if err != nil {
			return nil, err
		}
		sessionID := stringField(session, "sessionId")
		if options.OnSessionCreated != nil {
			options.OnSessionCreated(sessionID)
		}
		return c.CreatePlaintextConversationHandle(
			sessionID,
			encryptionSummaryFromResponse(session),
			options.Auth,
			options.UAID,
			"",
		), nil
	}

	handle, err := c.CreateEncryptedSession(ctx, StartEncryptedChatSessionOptions{
		UAID:              options.UAID,
		SenderUAID:        options.SenderUAID,
		HistoryTTLSeconds: options.HistoryTTLSeconds,
		HandshakeTimeout:  conversationTimeout(options.Encryption),
		PollInterval:      conversationPollInterval(options.Encryption),
		OnSessionCreated:  options.OnSessionCreated,
		Auth:              options.Auth,
	})
	if err == nil {
		return handle, nil
	}

	var unavailable *EncryptionUnavailableError
	if !errors.As(err, &unavailable) {
		return nil, err
	}
	if preference == "required" {
		return nil, err
	}

	return c.CreatePlaintextConversationHandle(
		unavailable.SessionID,
		unavailable.Summary,
		options.Auth,
		options.UAID,
		"",
	), nil
}

func (c *RegistryBrokerClient) AcceptConversation(
	ctx context.Context,
	options AcceptConversationOptions,
) (*ChatConversationHandle, error) {
	if err := ensureNonEmpty(options.SessionID, "sessionId"); err != nil {
		return nil, err
	}
	preference := conversationPreference(options.Encryption)
	if preference == "disabled" {
		return c.CreatePlaintextConversationHandle(
			options.SessionID,
			nil,
			nil,
			options.ResponderUAID,
			"",
		), nil
	}

	handle, err := c.AcceptEncryptedSession(ctx, AcceptEncryptedChatSessionOptions{
		SessionID:        options.SessionID,
		ResponderUAID:    options.ResponderUAID,
		HandshakeTimeout: conversationTimeout(options.Encryption),
		PollInterval:     conversationPollInterval(options.Encryption),
	})
	if err == nil {
		return handle, nil
	}

	var unavailable *EncryptionUnavailableError
	if !errors.As(err, &unavailable) {
		return nil, err
	}
	if preference == "required" {
		return nil, err
	}
	return c.CreatePlaintextConversationHandle(
		options.SessionID,
		unavailable.Summary,
		nil,
		options.ResponderUAID,
		"",
	), nil
}

func (c *RegistryBrokerClient) CompactHistory(
	ctx context.Context,
	payload CompactHistoryRequestPayload,
) (JSONObject, error) {
	if err := ensureNonEmpty(payload.SessionID, "sessionId"); err != nil {
		return nil, err
	}
	body := JSONObject{}
	if payload.PreserveEntries != nil && *payload.PreserveEntries >= 0 {
		body["preserveEntries"] = *payload.PreserveEntries
	}
	path := "/chat/session/" + percentPath(payload.SessionID) + "/compact"
	return c.requestJSON(
		ctx,
		http.MethodPost,
		path,
		body,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) FetchEncryptionStatus(
	ctx context.Context,
	sessionID string,
) (JSONObject, error) {
	if err := ensureNonEmpty(sessionID, "sessionId"); err != nil {
		return nil, err
	}
	path := "/chat/session/" + percentPath(sessionID) + "/encryption"
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

func (c *RegistryBrokerClient) PostEncryptionHandshake(
	ctx context.Context,
	sessionID string,
	payload EncryptionHandshakeSubmissionPayload,
) (JSONObject, error) {
	if err := ensureNonEmpty(sessionID, "sessionId"); err != nil {
		return nil, err
	}
	path := "/chat/session/" + percentPath(sessionID) + "/encryption-handshake"
	body := bodyMap(payload)
	result, err := c.requestJSON(
		ctx,
		http.MethodPost,
		path,
		body,
		map[string]string{"content-type": "application/json"},
	)
	if err != nil {
		return nil, err
	}
	if handshakeValue, ok := result["handshake"].(map[string]any); ok {
		return handshakeValue, nil
	}
	return result, nil
}

func (c *RegistryBrokerClient) SendMessage(
	ctx context.Context,
	payload SendMessageRequestPayload,
) (JSONObject, error) {
	message := strings.TrimSpace(payload.Message)
	cipherEnvelope := payload.CipherEnvelope
	if payload.Encryption != nil {
		sessionID := strings.TrimSpace(payload.Encryption.SessionID)
		if sessionID == "" {
			sessionID = strings.TrimSpace(payload.SessionID)
		}
		if sessionID == "" {
			return nil, fmt.Errorf("sessionId is required when using encrypted chat payloads")
		}
		if len(payload.Encryption.Recipients) == 0 {
			return nil, fmt.Errorf("recipients are required for encrypted chat payloads")
		}
		envelope, err := c.BuildCipherEnvelope(EncryptCipherEnvelopeOptions{
			Plaintext:      payload.Encryption.Plaintext,
			SessionID:      sessionID,
			Revision:       payload.Encryption.Revision,
			AssociatedData: payload.Encryption.AssociatedData,
			SharedSecret:   payload.Encryption.SharedSecret,
			Recipients:     payload.Encryption.Recipients,
		})
		if err != nil {
			return nil, err
		}
		cipherEnvelope = &envelope
		if strings.TrimSpace(payload.SessionID) == "" {
			payload.SessionID = sessionID
		}
		if message == "" {
			message = "[ciphertext omitted]"
		}
	}
	if err := ensureNonEmpty(message, "message"); err != nil {
		return nil, err
	}
	body := JSONObject{"message": message}
	addObjectString(body, "sessionId", payload.SessionID)
	addObjectString(body, "uaid", payload.UAID)
	addObjectString(body, "agentUrl", payload.AgentURL)
	if payload.Streaming != nil {
		body["streaming"] = *payload.Streaming
	}
	if payload.Auth != nil {
		body["auth"] = serialiseAuthConfig(payload.Auth)
	}
	if cipherEnvelope != nil {
		body["cipherEnvelope"] = bodyMap(*cipherEnvelope)
	}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/chat/message",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) EndSession(ctx context.Context, sessionID string) error {
	if err := ensureNonEmpty(sessionID, "sessionId"); err != nil {
		return err
	}
	path := "/chat/session/" + percentPath(sessionID)
	return c.requestNoResponse(ctx, http.MethodDelete, path, nil, nil)
}

func (c *RegistryBrokerClient) CreatePlaintextConversationHandle(
	sessionID string,
	summary JSONObject,
	defaultAuth *AgentAuthConfig,
	uaid string,
	agentURL string,
) *ChatConversationHandle {
	return &ChatConversationHandle{
		SessionID:    sessionID,
		Mode:         "plaintext",
		Summary:      summary,
		client:       c,
		defaultAuth:  defaultAuth,
		uaid:         strings.TrimSpace(uaid),
		agentURL:     strings.TrimSpace(agentURL),
		sharedSecret: nil,
		recipients:   nil,
		identity:     nil,
	}
}
