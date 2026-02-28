package inscriber

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
)

type InscriptionMode string

const (
	ModeFile               InscriptionMode = "file"
	ModeUpload             InscriptionMode = "upload"
	ModeHashinal           InscriptionMode = "hashinal"
	ModeHashinalCollection InscriptionMode = "hashinal-collection"
	ModeBulkFiles          InscriptionMode = "bulk-files"
)

type FileInput struct {
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Base64   string `json:"base64,omitempty"`
	FileName string `json:"fileName,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type ConnectionMode string

const (
	ConnectionModeHTTP      ConnectionMode = "http"
	ConnectionModeWebSocket ConnectionMode = "websocket"
	ConnectionModeAuto      ConnectionMode = "auto"
)

type InscriptionInputType string

const (
	InscriptionInputTypeURL    InscriptionInputType = "url"
	InscriptionInputTypeFile   InscriptionInputType = "file"
	InscriptionInputTypeBuffer InscriptionInputType = "buffer"
)

type InscriptionInput struct {
	Type     InscriptionInputType
	URL      string
	Path     string
	Buffer   []byte
	FileName string
	MimeType string
}

type RegistrationStage string

const (
	RegistrationStagePreparing  RegistrationStage = "preparing"
	RegistrationStageSubmitting RegistrationStage = "submitting"
	RegistrationStageConfirming RegistrationStage = "confirming"
	RegistrationStageVerifying  RegistrationStage = "verifying"
	RegistrationStageCompleted  RegistrationStage = "completed"
)

type RegistrationProgressData struct {
	Stage           RegistrationStage
	Message         string
	ProgressPercent float64
	Details         map[string]any
}

type RegistrationProgressCallback func(data RegistrationProgressData)

type InscriptionOptions struct {
	Mode                InscriptionMode
	WebSocket           *bool
	ConnectionMode      ConnectionMode
	WaitForConfirmation *bool
	WaitMaxAttempts     int
	WaitInterval        int64
	APIKey              string
	BaseURL             string
	Tags                []string
	Metadata            map[string]any
	JSONFileURL         string
	FileStandard        string
	ChunkSize           int
	Network             Network
	QuoteOnly           bool
	ProgressCallback    RegistrationProgressCallback
}

type QuoteTransfer struct {
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

type QuoteResult struct {
	TotalCostHBAR string `json:"totalCostHbar"`
	ValidUntil    string `json:"validUntil"`
	Breakdown     struct {
		Transfers []QuoteTransfer `json:"transfers"`
	} `json:"breakdown"`
}

type InscriptionCostSummary struct {
	TotalCostHBAR string `json:"totalCostHbar"`
	Breakdown     struct {
		Transfers []QuoteTransfer `json:"transfers"`
	} `json:"breakdown"`
	ValidUntil string `json:"validUntil,omitempty"`
}

type InscriptionResponse struct {
	Confirmed   bool                    `json:"confirmed"`
	Result      any                     `json:"result"`
	Inscription *InscriptionJob         `json:"inscription,omitempty"`
	Quote       bool                    `json:"quote,omitempty"`
	CostSummary *InscriptionCostSummary `json:"costSummary,omitempty"`
}

type RetrieveInscriptionOptions struct {
	APIKey     string
	AccountID  string
	PrivateKey string
	Network    Network
	BaseURL    string
}

type StartInscriptionRequest struct {
	File               FileInput       `json:"file"`
	HolderID           string          `json:"holderId"`
	TTL                int64           `json:"ttl,omitempty"`
	Mode               InscriptionMode `json:"mode"`
	Network            Network         `json:"network,omitempty"`
	Metadata           map[string]any  `json:"metadata,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	Creator            string          `json:"creator,omitempty"`
	Description        string          `json:"description,omitempty"`
	FileStandard       string          `json:"fileStandard,omitempty"`
	ChunkSize          int             `json:"chunkSize,omitempty"`
	OnlyJSONCollection bool            `json:"onlyJSONCollection,omitempty"`
	JSONFileURL        string          `json:"jsonFileURL,omitempty"`
	MetadataObject     map[string]any  `json:"metadataObject,omitempty"`
}

type HederaClientConfig struct {
	AccountID  string
	PrivateKey string
	Network    Network
}

type InscriptionJob struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	Completed        bool   `json:"completed"`
	TransactionID    string `json:"transactionId,omitempty"`
	TransactionBytes string `json:"transactionBytes,omitempty"`
	TxID             string `json:"tx_id,omitempty"`
	TopicID          string `json:"topic_id,omitempty"`
	Error            string `json:"error,omitempty"`
	TotalCost        int64  `json:"totalCost,omitempty"`
	TotalMessages    int64  `json:"totalMessages,omitempty"`
}

type InscriptionResult struct {
	JobID         string `json:"jobId"`
	TransactionID string `json:"transactionId"`
	TopicID       string `json:"topicId,omitempty"`
	Status        string `json:"status,omitempty"`
	Completed     bool   `json:"completed"`
}

type AuthResult struct {
	APIKey string `json:"apiKey"`
}

type BrokerQuoteRequest struct {
	InputType    string          `json:"inputType"`
	Mode         InscriptionMode `json:"mode"`
	URL          string          `json:"url,omitempty"`
	Base64       string          `json:"base64,omitempty"`
	FileName     string          `json:"fileName,omitempty"`
	MimeType     string          `json:"mimeType,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	FileStandard string          `json:"fileStandard,omitempty"`
	ChunkSize    int             `json:"chunkSize,omitempty"`
}

type BrokerQuoteResponse struct {
	QuoteID       string  `json:"quoteId"`
	ContentHash   string  `json:"contentHash"`
	SizeBytes     int64   `json:"sizeBytes"`
	TotalCostHBAR float64 `json:"totalCostHbar"`
	Credits       float64 `json:"credits"`
	USDCents      int64   `json:"usdCents"`
	ExpiresAt     string  `json:"expiresAt"`
	Mode          string  `json:"mode"`
}

type InscribeViaRegistryBrokerOptions struct {
	BaseURL             string
	LedgerAPIKey        string
	APIKey              string
	Mode                InscriptionMode
	Metadata            map[string]any
	Tags                []string
	FileStandard        string
	ChunkSize           int
	WaitForConfirmation *bool
	WaitTimeoutMs       int64
	PollIntervalMs      int64
}

type BrokerJobResponse struct {
	JobID       string  `json:"jobId"`
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	HRL         string  `json:"hrl"`
	TopicID     string  `json:"topicId"`
	Network     string  `json:"network"`
	Credits     float64 `json:"credits"`
	QuoteCredit float64 `json:"quoteCredits"`
	USDCents    int64   `json:"usdCents"`
	QuoteCents  int64   `json:"quoteUsdCents"`
	SizeBytes   int64   `json:"sizeBytes"`
	Error       string  `json:"error"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type InscribeViaBrokerResult struct {
	Confirmed bool   `json:"confirmed"`
	JobID     string `json:"jobId"`
	Status    string `json:"status"`
	HRL       string `json:"hrl,omitempty"`
	TopicID   string `json:"topicId,omitempty"`
	Network   string `json:"network,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}
