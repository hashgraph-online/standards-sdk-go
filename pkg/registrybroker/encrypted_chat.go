package registrybroker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultEncryptionHandshakeTimeout = 30 * time.Second
	DefaultEncryptionPollInterval     = time.Second
)

type EncryptionUnavailableError struct {
	SessionID string
	Summary   JSONObject
}

// Error performs the requested operation.
func (e *EncryptionUnavailableError) Error() string {
	return "encryption is not enabled for this session"
}

// CreateEncryptedSession creates the requested resource.
func (c *RegistryBrokerClient) CreateEncryptedSession(
	ctx context.Context,
	options StartEncryptedChatSessionOptions,
) (*ChatConversationHandle, error) {
	if err := ensureNonEmpty(options.UAID, "uaid"); err != nil {
		return nil, err
	}

	requestEncryption := true
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

	sessionID := strings.TrimSpace(stringField(session, "sessionId"))
	if options.OnSessionCreated != nil && sessionID != "" {
		options.OnSessionCreated(sessionID)
	}

	summary := encryptionSummaryFromResponse(session)
	if !encryptionEnabled(summary) {
		return nil, &EncryptionUnavailableError{
			SessionID: sessionID,
			Summary:   summary,
		}
	}

	return c.establishRequesterContext(ctx, requesterContextOptions{
		sessionID:        sessionID,
		summary:          summary,
		senderUAID:       options.SenderUAID,
		handshakeTimeout: normalizedTimeout(options.HandshakeTimeout),
		pollInterval:     normalizedPollInterval(options.PollInterval),
		defaultAuth:      options.Auth,
	})
}

// AcceptEncryptedSession performs the requested operation.
func (c *RegistryBrokerClient) AcceptEncryptedSession(
	ctx context.Context,
	options AcceptEncryptedChatSessionOptions,
) (*ChatConversationHandle, error) {
	if err := ensureNonEmpty(options.SessionID, "sessionId"); err != nil {
		return nil, err
	}

	summary, err := c.waitForEncryptionSummary(
		ctx,
		options.SessionID,
		normalizedTimeout(options.HandshakeTimeout),
		normalizedPollInterval(options.PollInterval),
	)
	if err != nil {
		return nil, err
	}

	return c.establishResponderContext(ctx, responderContextOptions{
		sessionID:        options.SessionID,
		summary:          summary,
		responderUAID:    options.ResponderUAID,
		handshakeTimeout: normalizedTimeout(options.HandshakeTimeout),
		pollInterval:     normalizedPollInterval(options.PollInterval),
	})
}

type requesterContextOptions struct {
	sessionID        string
	summary          JSONObject
	senderUAID       string
	handshakeTimeout time.Duration
	pollInterval     time.Duration
	defaultAuth      *AgentAuthConfig
}

type responderContextOptions struct {
	sessionID        string
	summary          JSONObject
	responderUAID    string
	handshakeTimeout time.Duration
	pollInterval     time.Duration
}

func (c *RegistryBrokerClient) establishRequesterContext(
	ctx context.Context,
	options requesterContextOptions,
) (*ChatConversationHandle, error) {
	keyPair, err := c.GenerateEphemeralKeyPair()
	if err != nil {
		return nil, err
	}

	requesterUAID := strings.TrimSpace(options.senderUAID)
	if requesterUAID == "" {
		requesterUAID = strings.TrimSpace(stringField(peerFromSummary(options.summary, "requester"), "uaid"))
	}

	_, err = c.PostEncryptionHandshake(ctx, options.sessionID, EncryptionHandshakeSubmissionPayload{
		Role:               "requester",
		KeyType:            "secp256k1",
		EphemeralPublicKey: keyPair.PublicKey,
		UAID:               requesterUAID,
	})
	if err != nil {
		return nil, err
	}

	summary, record, err := c.waitForHandshakeCompletion(
		ctx,
		options.sessionID,
		options.handshakeTimeout,
		options.pollInterval,
	)
	if err != nil {
		return nil, err
	}

	responderKey := strings.TrimSpace(stringField(peerFromSummary(record, "responder"), "ephemeralPublicKey"))
	if responderKey == "" {
		return nil, fmt.Errorf("responder handshake was not completed in time")
	}

	sharedSecret, err := c.DeriveSharedSecret(DeriveSharedSecretOptions{
		PrivateKey:    keyPair.PrivateKey,
		PeerPublicKey: responderKey,
	})
	if err != nil {
		return nil, err
	}

	return c.createEncryptedConversationHandle(
		options.sessionID,
		summary,
		sharedSecret,
		buildRecipientsFromSummary(summary),
		identityFromSummary(summary, "requester"),
		options.defaultAuth,
	)
}

func (c *RegistryBrokerClient) establishResponderContext(
	ctx context.Context,
	options responderContextOptions,
) (*ChatConversationHandle, error) {
	keyPair, err := c.GenerateEphemeralKeyPair()
	if err != nil {
		return nil, err
	}

	responderUAID := strings.TrimSpace(options.responderUAID)
	if responderUAID == "" {
		responderUAID = strings.TrimSpace(stringField(peerFromSummary(options.summary, "responder"), "uaid"))
	}

	_, err = c.PostEncryptionHandshake(ctx, options.sessionID, EncryptionHandshakeSubmissionPayload{
		Role:               "responder",
		KeyType:            "secp256k1",
		EphemeralPublicKey: keyPair.PublicKey,
		UAID:               responderUAID,
	})
	if err != nil {
		return nil, err
	}

	summary, record, err := c.waitForHandshakeCompletion(
		ctx,
		options.sessionID,
		options.handshakeTimeout,
		options.pollInterval,
	)
	if err != nil {
		return nil, err
	}

	requesterKey := strings.TrimSpace(stringField(peerFromSummary(record, "requester"), "ephemeralPublicKey"))
	if requesterKey == "" {
		return nil, fmt.Errorf("requester handshake was not detected in time")
	}

	sharedSecret, err := c.DeriveSharedSecret(DeriveSharedSecretOptions{
		PrivateKey:    keyPair.PrivateKey,
		PeerPublicKey: requesterKey,
	})
	if err != nil {
		return nil, err
	}

	return c.createEncryptedConversationHandle(
		options.sessionID,
		summary,
		sharedSecret,
		buildRecipientsFromSummary(summary),
		identityFromSummary(summary, "responder"),
		nil,
	)
}

func (c *RegistryBrokerClient) waitForHandshakeCompletion(
	ctx context.Context,
	sessionID string,
	timeout time.Duration,
	pollInterval time.Duration,
) (JSONObject, JSONObject, error) {
	deadline := time.Now().Add(timeout)
	for {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}

		status, err := c.FetchEncryptionStatus(ctx, sessionID)
		if err != nil {
			return nil, nil, err
		}
		summary := encryptionSummaryFromResponse(status)
		record := peerFromSummary(summary, "handshake")
		if strings.EqualFold(strings.TrimSpace(stringField(record, "status")), "complete") {
			return summary, record, nil
		}
		if time.Now().After(deadline) {
			return nil, nil, fmt.Errorf("timed out waiting for encrypted handshake completion")
		}
		if err := c.delay(ctx, pollInterval); err != nil {
			return nil, nil, err
		}
	}
}

func (c *RegistryBrokerClient) waitForEncryptionSummary(
	ctx context.Context,
	sessionID string,
	timeout time.Duration,
	pollInterval time.Duration,
) (JSONObject, error) {
	deadline := time.Now().Add(timeout)
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		status, err := c.FetchEncryptionStatus(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		summary := encryptionSummaryFromResponse(status)
		if encryptionEnabled(summary) {
			return summary, nil
		}
		if time.Now().After(deadline) {
			return nil, &EncryptionUnavailableError{
				SessionID: sessionID,
				Summary:   summary,
			}
		}
		if err := c.delay(ctx, pollInterval); err != nil {
			return nil, err
		}
	}
}

func (c *RegistryBrokerClient) createEncryptedConversationHandle(
	sessionID string,
	summary JSONObject,
	sharedSecret []byte,
	recipients []CipherEnvelopeRecipient,
	identity *RecipientIdentity,
	defaultAuth *AgentAuthConfig,
) (*ChatConversationHandle, error) {
	if err := ensureNonEmpty(sessionID, "sessionId"); err != nil {
		return nil, err
	}
	uaid := strings.TrimSpace(stringField(peerFromSummary(summary, "requester"), "uaid"))
	if uaid == "" {
		uaid = strings.TrimSpace(stringField(peerFromSummary(summary, "responder"), "uaid"))
	}
	if uaid == "" && identity != nil {
		uaid = strings.TrimSpace(identity.UAID)
	}

	handle := c.CreatePlaintextConversationHandle(sessionID, summary, defaultAuth, uaid, "")
	handle.Mode = "encrypted"
	handle.sharedSecret = cloneBytes(sharedSecret)
	handle.recipients = cloneRecipients(recipients)
	handle.identity = cloneRecipientIdentity(identity)
	c.RegisterConversationContextForEncryption(ConversationContextInput{
		SessionID:    sessionID,
		SharedSecret: sharedSecret,
		Identity:     cloneRecipientIdentity(identity),
	})
	return handle, nil
}

func conversationPreference(options *ConversationEncryptionOptions) string {
	if options == nil {
		return "preferred"
	}
	preference := strings.ToLower(strings.TrimSpace(options.Preference))
	if preference == "" {
		return "preferred"
	}
	return preference
}

func conversationTimeout(options *ConversationEncryptionOptions) time.Duration {
	if options == nil {
		return DefaultEncryptionHandshakeTimeout
	}
	return normalizedTimeout(options.HandshakeTimeout)
}

func conversationPollInterval(options *ConversationEncryptionOptions) time.Duration {
	if options == nil {
		return DefaultEncryptionPollInterval
	}
	return normalizedPollInterval(options.PollInterval)
}

func normalizedTimeout(value time.Duration) time.Duration {
	if value <= 0 {
		return DefaultEncryptionHandshakeTimeout
	}
	return value
}

func normalizedPollInterval(value time.Duration) time.Duration {
	if value <= 0 {
		return DefaultEncryptionPollInterval
	}
	return value
}

func encryptionSummaryFromResponse(response JSONObject) JSONObject {
	return peerFromSummary(response, "encryption")
}

func encryptionEnabled(summary JSONObject) bool {
	if summary == nil {
		return false
	}
	value, ok := summary["enabled"].(bool)
	return ok && value
}

func peerFromSummary(source JSONObject, key string) JSONObject {
	if source == nil {
		return nil
	}
	value, exists := source[key]
	if !exists || value == nil {
		return nil
	}
	if typed, ok := value.(JSONObject); ok {
		return typed
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func buildRecipientsFromSummary(summary JSONObject) []CipherEnvelopeRecipient {
	candidates := []JSONObject{
		peerFromSummary(summary, "requester"),
		peerFromSummary(summary, "responder"),
	}
	recipients := make([]CipherEnvelopeRecipient, 0, 2)
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		recipient := CipherEnvelopeRecipient{
			UAID:            strings.TrimSpace(stringField(candidate, "uaid")),
			LedgerAccountID: strings.TrimSpace(stringField(candidate, "ledgerAccountId")),
			UserID:          strings.TrimSpace(stringField(candidate, "userId")),
			Email:           strings.TrimSpace(stringField(candidate, "email")),
		}
		if recipient.UAID == "" &&
			recipient.LedgerAccountID == "" &&
			recipient.UserID == "" &&
			recipient.Email == "" {
			continue
		}
		recipients = append(recipients, recipient)
	}
	return recipients
}

func identityFromSummary(summary JSONObject, role string) *RecipientIdentity {
	peer := peerFromSummary(summary, role)
	if peer == nil {
		return nil
	}
	identity := &RecipientIdentity{
		UAID:            strings.TrimSpace(stringField(peer, "uaid")),
		LedgerAccountID: strings.TrimSpace(stringField(peer, "ledgerAccountId")),
		UserID:          strings.TrimSpace(stringField(peer, "userId")),
		Email:           strings.TrimSpace(stringField(peer, "email")),
	}
	if identity.UAID == "" &&
		identity.LedgerAccountID == "" &&
		identity.UserID == "" &&
		identity.Email == "" {
		return nil
	}
	return identity
}

func cloneRecipients(recipients []CipherEnvelopeRecipient) []CipherEnvelopeRecipient {
	if len(recipients) == 0 {
		return nil
	}
	cloned := make([]CipherEnvelopeRecipient, len(recipients))
	copy(cloned, recipients)
	return cloned
}
