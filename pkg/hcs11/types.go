package hcs11

type ProfileType int

const (
	ProfileTypePersonal  ProfileType = 0
	ProfileTypeAIAgent   ProfileType = 1
	ProfileTypeMCPServer ProfileType = 2
	ProfileTypeFlora     ProfileType = 3
)

type AIAgentType int

const (
	AIAgentTypeManual     AIAgentType = 0
	AIAgentTypeAutonomous AIAgentType = 1
)

type AIAgentCapability int

const (
	AIAgentCapabilityTextGeneration AIAgentCapability = iota
	AIAgentCapabilityImageGeneration
	AIAgentCapabilityAudioGeneration
	AIAgentCapabilityVideoGeneration
	AIAgentCapabilityCodeGeneration
	AIAgentCapabilityLanguageTranslation
	AIAgentCapabilitySummarizationExtraction
	AIAgentCapabilityKnowledgeRetrieval
	AIAgentCapabilityDataIntegration
	AIAgentCapabilityMarketIntelligence
	AIAgentCapabilityTransactionAnalytics
	AIAgentCapabilitySmartContractAudit
	AIAgentCapabilityGovernanceFacilitation
	AIAgentCapabilitySecurityMonitoring
	AIAgentCapabilityComplianceAnalysis
	AIAgentCapabilityFraudDetection
	AIAgentCapabilityMultiAgentCoordination
	AIAgentCapabilityAPIIntegration
	AIAgentCapabilityWorkflowAutomation
)

type MCPServerCapability int

const (
	MCPServerCapabilityResourceProvider MCPServerCapability = iota
	MCPServerCapabilityToolProvider
	MCPServerCapabilityPromptTemplateProvider
	MCPServerCapabilityLocalFileAccess
	MCPServerCapabilityDatabaseIntegration
	MCPServerCapabilityAPIIntegration
	MCPServerCapabilityWebAccess
	MCPServerCapabilityKnowledgeBase
	MCPServerCapabilityMemoryPersistence
	MCPServerCapabilityCodeAnalysis
	MCPServerCapabilityContentGeneration
	MCPServerCapabilityCommunication
	MCPServerCapabilityDocumentProcessing
	MCPServerCapabilityCalendarSchedule
	MCPServerCapabilitySearch
	MCPServerCapabilityAssistantOrchestration
)

type VerificationType string

const (
	VerificationTypeDNS       VerificationType = "dns"
	VerificationTypeSignature VerificationType = "signature"
	VerificationTypeChallenge VerificationType = "challenge"
)

type SocialPlatform string

const (
	SocialPlatformTwitter  SocialPlatform = "twitter"
	SocialPlatformGitHub   SocialPlatform = "github"
	SocialPlatformDiscord  SocialPlatform = "discord"
	SocialPlatformTelegram SocialPlatform = "telegram"
	SocialPlatformLinkedIn SocialPlatform = "linkedin"
	SocialPlatformYouTube  SocialPlatform = "youtube"
	SocialPlatformWebsite  SocialPlatform = "website"
	SocialPlatformX        SocialPlatform = "x"
)

type SocialLink struct {
	Platform SocialPlatform `json:"platform"`
	Handle   string         `json:"handle"`
}

type AIAgentDetails struct {
	Type         AIAgentType         `json:"type"`
	Capabilities []AIAgentCapability `json:"capabilities"`
	Model        string              `json:"model"`
	Creator      string              `json:"creator,omitempty"`
}

type MCPServerVerification struct {
	Type          VerificationType `json:"type"`
	Value         string           `json:"value"`
	DNSField      string           `json:"dns_field,omitempty"`
	ChallengePath string           `json:"challenge_path,omitempty"`
}

type MCPServerConnectionInfo struct {
	URL       string `json:"url"`
	Transport string `json:"transport"`
}

type MCPServerHost struct {
	MinVersion string `json:"minVersion,omitempty"`
}

type MCPServerResource struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MCPServerTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MCPServerDetails struct {
	Version        string                  `json:"version"`
	ConnectionInfo MCPServerConnectionInfo `json:"connectionInfo"`
	Services       []MCPServerCapability   `json:"services"`
	Description    string                  `json:"description"`
	Verification   *MCPServerVerification  `json:"verification,omitempty"`
	Host           *MCPServerHost          `json:"host,omitempty"`
	Capabilities   []string                `json:"capabilities,omitempty"`
	Resources      []MCPServerResource     `json:"resources,omitempty"`
	Tools          []MCPServerTool         `json:"tools,omitempty"`
	Maintainer     string                  `json:"maintainer,omitempty"`
	Repository     string                  `json:"repository,omitempty"`
	Docs           string                  `json:"docs,omitempty"`
}

type FloraMember struct {
	AccountID string `json:"accountId"`
	PublicKey string `json:"publicKey,omitempty"`
	Weight    int    `json:"weight,omitempty"`
}

type FloraTopics struct {
	Communication string `json:"communication"`
	Transaction   string `json:"transaction"`
	State         string `json:"state"`
}

type HCS11Profile struct {
	Version         string            `json:"version"`
	Type            ProfileType       `json:"type"`
	DisplayName     string            `json:"display_name"`
	Alias           string            `json:"alias,omitempty"`
	Bio             string            `json:"bio,omitempty"`
	Socials         []SocialLink      `json:"socials,omitempty"`
	ProfileImage    string            `json:"profileImage,omitempty"`
	UAID            string            `json:"uaid,omitempty"`
	Properties      map[string]any    `json:"properties,omitempty"`
	InboundTopicID  string            `json:"inboundTopicId,omitempty"`
	OutboundTopicID string            `json:"outboundTopicId,omitempty"`
	BaseAccount     string            `json:"base_account,omitempty"`
	AIAgent         *AIAgentDetails   `json:"aiAgent,omitempty"`
	MCPServer       *MCPServerDetails `json:"mcpServer,omitempty"`
	Members         []FloraMember     `json:"members,omitempty"`
	Threshold       int               `json:"threshold,omitempty"`
	Topics          *FloraTopics      `json:"topics,omitempty"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
	Policies        map[string]any    `json:"policies,omitempty"`
}

type Auth struct {
	OperatorID string
	PrivateKey string
}

type ClientConfig struct {
	Network           string
	Auth              Auth
	KeyType           string
	MirrorBaseURL     string
	KiloScribeBaseURL string
	InscriberAuthURL  string
	InscriberAPIURL   string
}

type ValidationResult struct {
	Valid  bool
	Errors []string
}

type TransactionResult struct {
	Success bool
	Error   string
}

type InscribeImageResponse struct {
	ImageTopicID  string
	TransactionID string
	Success       bool
	Error         string
	TotalCostHBAR string
}

type InscribeProfileResponse struct {
	ProfileTopicID  string
	TransactionID   string
	Success         bool
	Error           string
	InboundTopicID  string
	OutboundTopicID string
	TotalCostHBAR   string
}

type InscribeImageOptions struct {
	WaitForConfirmation bool
}

type InscribeProfileOptions struct {
	WaitForConfirmation bool
}

type FetchProfileResponse struct {
	Success   bool
	Profile   *HCS11Profile
	Error     string
	TopicInfo *ResolvedTopicInfo
}

type ResolvedTopicInfo struct {
	InboundTopic   string
	OutboundTopic  string
	ProfileTopicID string
}

type AgentMetadata struct {
	Type       string
	Model      string
	Socials    map[SocialPlatform]string
	Creator    string
	Properties map[string]any
}

var CapabilityNameToCapabilityMap = map[string]AIAgentCapability{
	"text_generation":          AIAgentCapabilityTextGeneration,
	"image_generation":         AIAgentCapabilityImageGeneration,
	"audio_generation":         AIAgentCapabilityAudioGeneration,
	"video_generation":         AIAgentCapabilityVideoGeneration,
	"code_generation":          AIAgentCapabilityCodeGeneration,
	"language_translation":     AIAgentCapabilityLanguageTranslation,
	"summarization_extraction": AIAgentCapabilitySummarizationExtraction,
	"knowledge_retrieval":      AIAgentCapabilityKnowledgeRetrieval,
	"data_integration":         AIAgentCapabilityDataIntegration,
	"market_intelligence":      AIAgentCapabilityMarketIntelligence,
	"transaction_analytics":    AIAgentCapabilityTransactionAnalytics,
	"smart_contract_audit":     AIAgentCapabilitySmartContractAudit,
	"governance_facilitation":  AIAgentCapabilityGovernanceFacilitation,
	"security_monitoring":      AIAgentCapabilitySecurityMonitoring,
	"compliance_analysis":      AIAgentCapabilityComplianceAnalysis,
	"fraud_detection":          AIAgentCapabilityFraudDetection,
	"multi_agent_coordination": AIAgentCapabilityMultiAgentCoordination,
	"api_integration":          AIAgentCapabilityAPIIntegration,
	"workflow_automation":      AIAgentCapabilityWorkflowAutomation,
}
