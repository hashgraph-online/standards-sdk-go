package registrybroker

import (
	"context"
	"net/http"
	"time"
)

type JSONObject map[string]any

type RegistryBrokerClientOptions struct {
	BaseURL             string
	APIKey              string
	AccountID           string
	LedgerAPIKey        string
	DefaultHeaders      map[string]string
	HTTPClient          *http.Client
	HTTPTimeout         time.Duration
	RegistrationAutoTop *AutoTopUpOptions
	HistoryAutoTop      *HistoryAutoTopUpOptions
}

type AutoTopUpOptions struct {
	AccountID  string
	PrivateKey string
	Memo       string
}

type HistoryAutoTopUpOptions struct {
	AccountID  string
	PrivateKey string
	HbarAmount float64
	Memo       string
}

type RegisterAgentOptions struct {
	AutoTopUp *AutoTopUpOptions
}

type RegistrationProgressWaitOptions struct {
	Interval       time.Duration
	Timeout        time.Duration
	ThrowOnFailure *bool
	OnProgress     func(JSONObject)
}

type SearchParams struct {
	Q            string
	Page         int
	Limit        int
	Registry     string
	Registries   []string
	MinTrust     *float64
	Capabilities []string
	Protocols    []string
	Adapters     []string
	Metadata     map[string][]any
	Type         string
	Verified     *bool
	Online       *bool
	SortBy       string
	SortOrder    string
}

type VectorSearchFilter struct {
	Registry     string   `json:"registry,omitempty"`
	Protocols    []string `json:"protocols,omitempty"`
	Adapter      []string `json:"adapter,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Type         string   `json:"type,omitempty"`
}

type VectorSearchRequest struct {
	Query  string              `json:"query"`
	Limit  int                 `json:"limit,omitempty"`
	Offset int                 `json:"offset,omitempty"`
	Filter *VectorSearchFilter `json:"filter,omitempty"`
}

type AgentAuthConfig struct {
	Type        string            `json:"type,omitempty"`
	Token       string            `json:"token,omitempty"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	HeaderName  string            `json:"headerName,omitempty"`
	HeaderValue string            `json:"headerValue,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type StartChatOptions struct {
	UAID              string
	AgentURL          string
	SenderUAID        string
	HistoryTTLSeconds *int
	Auth              *AgentAuthConfig
	Encryption        *ConversationEncryptionOptions
	OnSessionCreated  func(string)
}

type StartConversationOptions struct {
	UAID              string
	SenderUAID        string
	HistoryTTLSeconds *int
	Auth              *AgentAuthConfig
	Encryption        *ConversationEncryptionOptions
	OnSessionCreated  func(string)
}

type AcceptConversationOptions struct {
	SessionID     string
	ResponderUAID string
	Encryption    *ConversationEncryptionOptions
}

type ConversationEncryptionOptions struct {
	Preference       string
	HandshakeTimeout time.Duration
	PollInterval     time.Duration
}

type StartEncryptedChatSessionOptions struct {
	UAID              string
	SenderUAID        string
	HistoryTTLSeconds *int
	HandshakeTimeout  time.Duration
	PollInterval      time.Duration
	OnSessionCreated  func(string)
	Auth              *AgentAuthConfig
}

type AcceptEncryptedChatSessionOptions struct {
	SessionID        string
	ResponderUAID    string
	HandshakeTimeout time.Duration
	PollInterval     time.Duration
}

type SendMessageEncryptionOptions struct {
	Plaintext      string
	SessionID      string
	Revision       int
	AssociatedData string
	SharedSecret   []byte
	Recipients     []CipherEnvelopeRecipient
}

type CreateSessionRequestPayload struct {
	UAID                string
	AgentURL            string
	SenderUAID          string
	HistoryTTLSeconds   *int
	EncryptionRequested *bool
	Auth                *AgentAuthConfig
}

type SendMessageRequestPayload struct {
	Message        string
	SessionID      string
	UAID           string
	AgentURL       string
	Streaming      *bool
	Auth           *AgentAuthConfig
	CipherEnvelope *CipherEnvelope
	Encryption     *SendMessageEncryptionOptions
}

type CompactHistoryRequestPayload struct {
	SessionID       string
	PreserveEntries *int
}

type EncryptionHandshakeSubmissionPayload struct {
	Role               string     `json:"role,omitempty"`
	KeyType            string     `json:"keyType,omitempty"`
	EphemeralPublicKey string     `json:"ephemeralPublicKey,omitempty"`
	LongTermPublicKey  string     `json:"longTermPublicKey,omitempty"`
	Signature          string     `json:"signature,omitempty"`
	UAID               string     `json:"uaid,omitempty"`
	UserID             string     `json:"userId,omitempty"`
	LedgerAccountID    string     `json:"ledgerAccountId,omitempty"`
	Metadata           JSONObject `json:"metadata,omitempty"`
}

type ChatHistoryFetchOptions struct {
	Decrypt      *bool
	SharedSecret []byte
	Identity     *RecipientIdentity
}

type RecipientIdentity struct {
	UAID            string
	LedgerAccountID string
	UserID          string
	Email           string
}

type ConversationContextInput struct {
	SessionID    string
	SharedSecret []byte
	Identity     *RecipientIdentity
}

type ConversationContextState struct {
	SessionID    string
	SharedSecret []byte
	Identity     *RecipientIdentity
}

type DecryptedHistoryEntry struct {
	Entry     JSONObject
	Plaintext *string
}

type ChatConversationHandle struct {
	SessionID    string
	Mode         string
	Summary      JSONObject
	client       *RegistryBrokerClient
	defaultAuth  *AgentAuthConfig
	uaid         string
	agentURL     string
	sharedSecret []byte
	recipients   []CipherEnvelopeRecipient
	identity     *RecipientIdentity
}

type LedgerChallengeRequest struct {
	AccountID string
	Network   string
}

type LedgerVerifyRequest struct {
	ChallengeID      string
	AccountID        string
	Network          string
	Signature        string
	SignatureKind    string
	PublicKey        string
	ExpiresInMinutes *int
}

type LedgerAuthenticationSignerResult struct {
	Signature     string
	SignatureKind string
	PublicKey     string
}

type LedgerSignFunc func(message string) (LedgerAuthenticationSignerResult, error)

type LedgerAuthenticationOptions struct {
	AccountID        string
	Network          string
	ExpiresInMinutes *int
	Sign             LedgerSignFunc
}

type LedgerCredentialAuthOptions struct {
	AccountID        string
	Network          string
	ExpiresInMinutes *int
	Signature        string
	SignatureKind    string
	PublicKey        string
	Sign             SignMessageFunc
	SetAccountHeader bool
}

type AgentFeedbackQuery struct {
	IncludeRevoked bool
}

type AgentFeedbackIndexOptions struct {
	Page       *int
	Limit      *int
	Registries []string
}

type AdapterRegistryFilters struct {
	Category string
	Entity   string
	Keywords []string
	Query    string
	Limit    *int
	Offset   *int
}

type SkillCatalogOptions struct {
	Q        string
	Category string
	Tags     []string
	Featured *bool
	Verified *bool
	Channel  string
	SortBy   string
	Limit    *int
	Cursor   string
}

type ListSkillsOptions struct {
	Name         string
	Version      string
	Limit        *int
	Cursor       string
	IncludeFiles *bool
	AccountID    string
}

type ListMySkillsOptions struct {
	Limit *int
}

type MySkillsListOptions struct {
	Limit     *int
	Cursor    string
	AccountID string
}

type SkillPublishJobOptions struct {
	AccountID string
}

type PurchaseCreditsWithHbarParams struct {
	AccountID  string
	PrivateKey string
	HbarAmount float64
	Memo       string
	Metadata   JSONObject
}

type PurchaseCreditsWithX402Params struct {
	AccountID      string
	Credits        int
	USDAmount      *float64
	Description    string
	Metadata       JSONObject
	PaymentHeaders map[string]string
}

type BuyCreditsWithX402Params struct {
	AccountID     string
	Credits       int
	USDAmount     *float64
	Description   string
	Metadata      JSONObject
	EVMPrivateKey string
	Network       string
	RPCURL        string
}

type RegisterEncryptionKeyPayload struct {
	KeyType         string `json:"keyType,omitempty"`
	PublicKey       string `json:"publicKey,omitempty"`
	UAID            string `json:"uaid,omitempty"`
	LedgerAccountID string `json:"ledgerAccountId,omitempty"`
	LedgerNetwork   string `json:"ledgerNetwork,omitempty"`
	Email           string `json:"email,omitempty"`
}

type EnsureAgentKeyOptions struct {
	UAID              string
	KeyType           string
	PublicKey         string
	PrivateKey        string
	LedgerAccountID   string
	LedgerNetwork     string
	Email             string
	GenerateIfMissing bool
}

type EphemeralKeyPair struct {
	PrivateKey string
	PublicKey  string
}

type DeriveSharedSecretOptions struct {
	PrivateKey    string
	PeerPublicKey string
}

type CipherEnvelopeRecipient struct {
	UAID            string `json:"uaid,omitempty"`
	UserID          string `json:"userId,omitempty"`
	LedgerAccountID string `json:"ledgerAccountId,omitempty"`
	Email           string `json:"email,omitempty"`
	EncryptedShare  string `json:"encryptedShare,omitempty"`
}

type CipherEnvelope struct {
	Algorithm      string                    `json:"algorithm"`
	Ciphertext     string                    `json:"ciphertext"`
	Nonce          string                    `json:"nonce"`
	AssociatedData string                    `json:"associatedData,omitempty"`
	KeyLocator     JSONObject                `json:"keyLocator"`
	Recipients     []CipherEnvelopeRecipient `json:"recipients,omitempty"`
}

type EncryptCipherEnvelopeOptions struct {
	Plaintext      string
	SessionID      string
	Revision       int
	AssociatedData string
	SharedSecret   []byte
	Recipients     []CipherEnvelopeRecipient
}

type DecryptCipherEnvelopeOptions struct {
	Envelope     CipherEnvelope
	SharedSecret []byte
}

type CreateAdapterRegistryCategoryRequest = JSONObject
type SubmitAdapterRegistryAdapterRequest = JSONObject
type AgentRegistrationRequest = JSONObject
type MoltbookOwnerRegistrationUpdateRequest = JSONObject
type AgentFeedbackEligibilityRequest = JSONObject
type AgentFeedbackSubmissionRequest = JSONObject
type SkillRegistryQuoteRequest = JSONObject
type SkillRegistryPublishRequest = JSONObject
type SkillRecommendedVersionSetRequest = JSONObject
type SkillDeprecationSetRequest = JSONObject
type SkillRegistryVoteRequest = JSONObject
type SkillVerificationRequestCreateRequest = JSONObject
type VerifyVerificationChallengeRequest = JSONObject

type InitializeAgentClientOptions struct {
	RegistryBrokerClientOptions
	UAID                    string
	EnsureEncryptionKey     *bool
	EnsureEncryptionOptions *EnsureAgentKeyOptions
}

type InitializedAgentClient struct {
	Client     *RegistryBrokerClient
	Encryption JSONObject
}

type SignMessageFunc func(ctx context.Context, message string) (LedgerAuthenticationSignerResult, error)
