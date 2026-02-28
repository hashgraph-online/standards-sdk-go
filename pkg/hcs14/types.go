package hcs14

import "context"

type CanonicalAgentData struct {
	Registry string
	Name     string
	Version  string
	Protocol string
	NativeID string
	Skills   []int
}

type RoutingParams struct {
	UID      string
	Registry string
	Proto    string
	NativeID string
	Domain   string
	Src      string
	Version  string
}

type ParsedUAID struct {
	Target string
	ID     string
	Params map[string]string
}

type UAIDMetadata struct {
	Profile                 string
	Resolved                bool
	VerificationLevel       string
	VerificationMethod      string
	PrecedenceSource        string
	ResolutionMode          string
	SelectedFollowupProfile string
	ReconstructedUAID       string
	BaseDID                 string
	BaseDIDResolved         bool
	Protocol                string
	Endpoint                string
	AgentCardURL            string
}

type UAIDResolutionError struct {
	Code    string
	Message string
	Details map[string]any
}

type UAIDResolutionResult struct {
	ID                 string
	DID                string
	AlsoKnownAs        []string
	VerificationMethod []DIDVerificationMethod
	Authentication     []string
	AssertionMethod    []string
	Service            []ServiceEndpoint
	Metadata           UAIDMetadata
	Error              *UAIDResolutionError
}

type ServiceEndpoint struct {
	ID              string
	Type            string
	ServiceEndpoint string
	Source          string
}

type DIDVerificationMethod struct {
	ID                  string
	Type                string
	Controller          string
	PublicKeyMultibase  string
	BlockchainAccountID string
}

type DIDDocument struct {
	ID                 string
	AlsoKnownAs        []string
	VerificationMethod []DIDVerificationMethod
	Authentication     []string
	AssertionMethod    []string
	Service            []ServiceEndpoint
}

type DIDResolver interface {
	Supports(did string) bool
	Resolve(ctx context.Context, did string) (*DIDDocument, error)
}

type DIDProfileResolver interface {
	ResolveProfile(
		ctx context.Context,
		did string,
		context DIDProfileResolverContext,
	) (*UAIDResolutionResult, error)
}

type UAIDProfileResolver interface {
	ProfileID() string
	Supports(uaid string, parsed ParsedUAID) bool
	ResolveProfile(
		ctx context.Context,
		uaid string,
		context UAIDProfileResolverContext,
	) (*UAIDResolutionResult, error)
}

type DIDProfileResolverContext struct {
	UAID        string
	ParsedUAID  *ParsedUAID
	DIDDocument *DIDDocument
}

type UAIDProfileResolverContext struct {
	ParsedUAID             ParsedUAID
	DID                    string
	DIDDocument            *DIDDocument
	ResolveDID             func(ctx context.Context, did string) (*DIDDocument, error)
	ResolveDIDProfile      func(ctx context.Context, did string, context DIDProfileResolverContext) (*UAIDResolutionResult, error)
	ResolveUaidProfileByID func(ctx context.Context, profileID string, uaid string) (*UAIDResolutionResult, error)
}

type DNSLookupFunc func(ctx context.Context, hostname string) ([]string, error)

type FollowupResolverFunc func(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error)
