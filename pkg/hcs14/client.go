package hcs14

import "context"

type Client struct {
	registry        *ResolverRegistry
	uaidDNSResolver *UaidDNSWebResolver
	ansDNSResolver  *ANSDNSWebResolver
	aidDNSResolver  *AIDDNSWebResolver
	uaidDIDResolver *UAIDDidResolutionResolver
}

type ClientOptions struct {
	DNSLookup DNSLookupFunc
	HTTP      ANSDNSWebResolverOptions
	UAID      UaidDNSWebResolverOptions
	AID       AIDDNSWebResolverOptions
	Registry  *ResolverRegistry
}

func NewClient(options ClientOptions) *Client {
	uaidOptions := options.UAID
	ansOptions := options.HTTP
	aidOptions := options.AID

	if options.DNSLookup != nil {
		uaidOptions.DNSLookup = options.DNSLookup
		ansOptions.DNSLookup = options.DNSLookup
		aidOptions.DNSLookup = options.DNSLookup
	}

	if !uaidOptions.RequireFullResolution {
		uaidOptions.EnableFollowupResolution = true
	}

	registry := options.Registry
	if registry == nil {
		registry = NewResolverRegistry()
	}

	uaidResolver := NewUaidDNSWebResolver(uaidOptions)
	ansResolver := NewANSDNSWebResolver(ansOptions)
	aidResolver := NewAIDDNSWebResolver(aidOptions)
	uaidDidResolver := NewUAIDDidResolutionResolver()

	registry.RegisterUAIDProfileResolver(uaidResolver)
	registry.RegisterUAIDProfileResolver(ansResolver)
	registry.RegisterUAIDProfileResolver(aidResolver)
	registry.RegisterUAIDProfileResolver(uaidDidResolver)

	return &Client{
		registry:        registry,
		uaidDNSResolver: uaidResolver,
		ansDNSResolver:  ansResolver,
		aidDNSResolver:  aidResolver,
		uaidDIDResolver: uaidDidResolver,
	}
}

func (client *Client) Resolve(ctx context.Context, uaid string) (*UAIDResolutionResult, error) {
	result, err := client.registry.ResolveUAIDProfile(ctx, uaid, ResolveUaidProfileOptions{})
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	return &UAIDResolutionResult{
		ID: uaid,
		Metadata: UAIDMetadata{
			Profile:  UAIDDNSWebProfileID,
			Resolved: false,
		},
		Error: &UAIDResolutionError{
			Code:    "ERR_NOT_APPLICABLE",
			Message: "no supported HCS-14 profile matched for this UAID",
		},
	}, nil
}

func (client *Client) ResolveProfile(
	ctx context.Context,
	uaid string,
	profileID string,
) (*UAIDResolutionResult, error) {
	return client.registry.ResolveUAIDProfile(ctx, uaid, ResolveUaidProfileOptions{
		ProfileID: profileID,
	})
}

func (client *Client) ResolveUAIDDNSWeb(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error) {
	return client.ResolveProfile(ctx, uaid, UAIDDNSWebProfileID)
}

func (client *Client) ResolveANSDNSWeb(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error) {
	return client.ResolveProfile(ctx, uaid, ANSDNSWebProfileID)
}

func (client *Client) ResolveAIDDNSWeb(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error) {
	return client.ResolveProfile(ctx, uaid, AIDDNSWebProfileID)
}

func (client *Client) ResolveUAIDDidResolution(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error) {
	return client.ResolveProfile(ctx, uaid, UAIDDidResolutionProfileID)
}

func (client *Client) Registry() *ResolverRegistry {
	return client.registry
}

func (client *Client) UaidDNSResolver() *UaidDNSWebResolver {
	return client.uaidDNSResolver
}

func (client *Client) AnsDNSResolver() *ANSDNSWebResolver {
	return client.ansDNSResolver
}

func (client *Client) AidDNSResolver() *AIDDNSWebResolver {
	return client.aidDNSResolver
}

func (client *Client) UaidDidResolver() *UAIDDidResolutionResolver {
	return client.uaidDIDResolver
}
