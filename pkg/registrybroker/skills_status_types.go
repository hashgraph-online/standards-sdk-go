package registrybroker

type SkillTrustTier string

type SkillStatusRequest struct {
	Name    string
	Version string
}

type SkillPreviewLookupRequest struct {
	Name    string
	Version string
}

type SkillPreviewByRepoRequest struct {
	Repo     string
	SkillDir string
	Ref      string
}

type SkillStatusNextStep struct {
	Kind        string  `json:"kind"`
	Priority    int     `json:"priority"`
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	URL         *string `json:"url,omitempty"`
	Href        *string `json:"href,omitempty"`
	Command     *string `json:"command,omitempty"`
}

type SkillPreviewSuggestedNextStep struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Command     *string `json:"command,omitempty"`
	Href        *string `json:"href,omitempty"`
}

type SkillPreviewReport struct {
	SchemaVersion      string                          `json:"schema_version"`
	ToolVersion        string                          `json:"tool_version"`
	PreviewID          string                          `json:"preview_id"`
	RepoURL            string                          `json:"repo_url"`
	RepoOwner          string                          `json:"repo_owner"`
	RepoName           string                          `json:"repo_name"`
	DefaultBranch      string                          `json:"default_branch"`
	CommitSHA          string                          `json:"commit_sha"`
	Ref                string                          `json:"ref"`
	EventName          string                          `json:"event_name"`
	WorkflowRunURL     string                          `json:"workflow_run_url"`
	SkillDir           string                          `json:"skill_dir"`
	Name               string                          `json:"name"`
	Version            string                          `json:"version"`
	ValidationStatus   string                          `json:"validation_status"`
	Findings           []any                           `json:"findings"`
	PackageSummary     JSONObject                      `json:"package_summary"`
	SuggestedNextSteps []SkillPreviewSuggestedNextStep `json:"suggested_next_steps"`
	GeneratedAt        string                          `json:"generated_at"`
}

type SkillPreviewRecord struct {
	ID            string             `json:"id"`
	PreviewID     string             `json:"previewId"`
	Source        string             `json:"source"`
	Report        SkillPreviewReport `json:"report"`
	GeneratedAt   string             `json:"generatedAt"`
	ExpiresAt     string             `json:"expiresAt"`
	StatusURL     string             `json:"statusUrl"`
	Authoritative bool               `json:"authoritative"`
}

type SkillPreviewLookupResponse struct {
	Found         bool                `json:"found"`
	Authoritative bool                `json:"authoritative"`
	Preview       *SkillPreviewRecord `json:"preview"`
	StatusURL     *string             `json:"statusUrl"`
	ExpiresAt     *string             `json:"expiresAt"`
}

type SkillStatusPreviewMetadata struct {
	PreviewID   string `json:"previewId"`
	RepoURL     string `json:"repoUrl"`
	RepoOwner   string `json:"repoOwner"`
	RepoName    string `json:"repoName"`
	CommitSHA   string `json:"commitSha"`
	Ref         string `json:"ref"`
	EventName   string `json:"eventName"`
	SkillDir    string `json:"skillDir"`
	GeneratedAt string `json:"generatedAt"`
	ExpiresAt   string `json:"expiresAt"`
	StatusURL   string `json:"statusUrl"`
}

type SkillStatusChecks struct {
	RepoCommitIntegrity bool `json:"repoCommitIntegrity"`
	ManifestIntegrity   bool `json:"manifestIntegrity"`
	DomainProof         bool `json:"domainProof"`
}

type SkillStatusVerificationSignals struct {
	PublisherBound   bool `json:"publisherBound"`
	DomainProof      bool `json:"domainProof"`
	VerifiedDomain   bool `json:"verifiedDomain"`
	PreviewValidated bool `json:"previewValidated"`
}

type SkillStatusProvenanceSignals struct {
	RepoCommitIntegrity  bool `json:"repoCommitIntegrity"`
	ManifestIntegrity    bool `json:"manifestIntegrity"`
	CanonicalRelease     bool `json:"canonicalRelease"`
	PreviewAvailable     bool `json:"previewAvailable"`
	PreviewAuthoritative bool `json:"previewAuthoritative"`
}

type SkillStatusResponse struct {
	Name                string                         `json:"name"`
	Version             *string                        `json:"version"`
	Published           bool                           `json:"published"`
	VerifiedDomain      bool                           `json:"verifiedDomain"`
	TrustTier           SkillTrustTier                 `json:"trustTier"`
	BadgeMetric         string                         `json:"badgeMetric"`
	Checks              SkillStatusChecks              `json:"checks"`
	NextSteps           []SkillStatusNextStep          `json:"nextSteps"`
	VerificationSignals SkillStatusVerificationSignals `json:"verificationSignals"`
	ProvenanceSignals   SkillStatusProvenanceSignals   `json:"provenanceSignals"`
	Publisher           *SkillPublisherMetadata        `json:"publisher"`
	Preview             *SkillStatusPreviewMetadata    `json:"preview,omitempty"`
	StatusURL           *string                        `json:"statusUrl,omitempty"`
}

type SkillQuotePreviewRequest struct {
	FileCount  int    `json:"fileCount"`
	TotalBytes int    `json:"totalBytes"`
	Name       string `json:"name,omitempty"`
	Version    string `json:"version,omitempty"`
	RepoURL    string `json:"repoUrl,omitempty"`
	SkillDir   string `json:"skillDir,omitempty"`
}

type SkillQuotePreviewRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type SkillQuotePreviewResponse struct {
	EstimatedCredits SkillQuotePreviewRange `json:"estimatedCredits"`
	EstimatedHbar    JSONObject             `json:"estimatedHbar,omitempty"`
	PricingVersion   string                 `json:"pricingVersion"`
	Assumptions      []string               `json:"assumptions,omitempty"`
	PurchaseURL      string                 `json:"purchaseUrl,omitempty"`
	PublishURL       string                 `json:"publishUrl,omitempty"`
	VerificationURL  string                 `json:"verificationUrl,omitempty"`
}

type SkillConversionSignalsResponse struct {
	RepoURL                     string                `json:"repoUrl"`
	SkillDir                    string                `json:"skillDir"`
	TrustTier                   SkillTrustTier        `json:"trustTier"`
	ActionInstalled             bool                  `json:"actionInstalled"`
	PreviewUploaded             bool                  `json:"previewUploaded"`
	PreviewID                   string                `json:"previewId,omitempty"`
	LastValidateSuccessAt       string                `json:"lastValidateSuccessAt,omitempty"`
	StalePreviewAgeDays         int                   `json:"stalePreviewAgeDays,omitempty"`
	Published                   bool                  `json:"published"`
	Verified                    bool                  `json:"verified"`
	PublishReady                bool                  `json:"publishReady"`
	PublishBlockedByMissingAuth bool                  `json:"publishBlockedByMissingAuth"`
	StatusURL                   string                `json:"statusUrl,omitempty"`
	PurchaseURL                 string                `json:"purchaseUrl,omitempty"`
	PublishURL                  string                `json:"publishUrl,omitempty"`
	VerificationURL             string                `json:"verificationUrl,omitempty"`
	NextSteps                   []SkillStatusNextStep `json:"nextSteps,omitempty"`
}
