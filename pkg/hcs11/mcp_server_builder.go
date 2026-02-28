package hcs11

import (
	"fmt"
	"strings"
)

type MCPServerBuilder struct {
	profile HCS11Profile
}

// NewMCPServerBuilder creates a new MCPServerBuilder.
func NewMCPServerBuilder() *MCPServerBuilder {
	return &MCPServerBuilder{
		profile: HCS11Profile{
			Version: "1.0",
			Type:    ProfileTypeMCPServer,
			MCPServer: &MCPServerDetails{
				ConnectionInfo: MCPServerConnectionInfo{},
				Services:       []MCPServerCapability{},
			},
		},
	}
}

// SetName sets the requested value.
func (builder *MCPServerBuilder) SetName(name string) *MCPServerBuilder {
	builder.profile.DisplayName = strings.TrimSpace(name)
	return builder
}

// SetAlias sets the requested value.
func (builder *MCPServerBuilder) SetAlias(alias string) *MCPServerBuilder {
	builder.profile.Alias = strings.TrimSpace(alias)
	return builder
}

// SetBio sets the requested value.
func (builder *MCPServerBuilder) SetBio(bio string) *MCPServerBuilder {
	builder.profile.Bio = strings.TrimSpace(bio)
	return builder
}

// SetDescription sets the requested value.
func (builder *MCPServerBuilder) SetDescription(description string) *MCPServerBuilder {
	return builder.SetBio(description)
}

// SetVersion sets the requested value.
func (builder *MCPServerBuilder) SetVersion(version string) *MCPServerBuilder {
	builder.profile.MCPServer.Version = strings.TrimSpace(version)
	return builder
}

// SetConnectionInfo sets the requested value.
func (builder *MCPServerBuilder) SetConnectionInfo(rawURL string, transport string) *MCPServerBuilder {
	builder.profile.MCPServer.ConnectionInfo = MCPServerConnectionInfo{
		URL:       strings.TrimSpace(rawURL),
		Transport: strings.TrimSpace(transport),
	}
	return builder
}

// SetServerDescription sets the requested value.
func (builder *MCPServerBuilder) SetServerDescription(description string) *MCPServerBuilder {
	builder.profile.MCPServer.Description = strings.TrimSpace(description)
	return builder
}

// SetServices sets the requested value.
func (builder *MCPServerBuilder) SetServices(services []MCPServerCapability) *MCPServerBuilder {
	builder.profile.MCPServer.Services = append([]MCPServerCapability{}, services...)
	return builder
}

// SetHostRequirements sets the requested value.
func (builder *MCPServerBuilder) SetHostRequirements(minVersion string) *MCPServerBuilder {
	builder.profile.MCPServer.Host = &MCPServerHost{MinVersion: strings.TrimSpace(minVersion)}
	return builder
}

// SetCapabilities sets the requested value.
func (builder *MCPServerBuilder) SetCapabilities(capabilities []string) *MCPServerBuilder {
	builder.profile.MCPServer.Capabilities = append([]string{}, capabilities...)
	return builder
}

// AddResource adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddResource(name string, description string) *MCPServerBuilder {
	builder.profile.MCPServer.Resources = append(builder.profile.MCPServer.Resources, MCPServerResource{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
	})
	return builder
}

// SetResources sets the requested value.
func (builder *MCPServerBuilder) SetResources(resources []MCPServerResource) *MCPServerBuilder {
	builder.profile.MCPServer.Resources = append([]MCPServerResource{}, resources...)
	return builder
}

// AddTool adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddTool(name string, description string) *MCPServerBuilder {
	builder.profile.MCPServer.Tools = append(builder.profile.MCPServer.Tools, MCPServerTool{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
	})
	return builder
}

// SetTools sets the requested value.
func (builder *MCPServerBuilder) SetTools(tools []MCPServerTool) *MCPServerBuilder {
	builder.profile.MCPServer.Tools = append([]MCPServerTool{}, tools...)
	return builder
}

// SetMaintainer sets the requested value.
func (builder *MCPServerBuilder) SetMaintainer(maintainer string) *MCPServerBuilder {
	builder.profile.MCPServer.Maintainer = strings.TrimSpace(maintainer)
	return builder
}

// SetRepository sets the requested value.
func (builder *MCPServerBuilder) SetRepository(repository string) *MCPServerBuilder {
	builder.profile.MCPServer.Repository = strings.TrimSpace(repository)
	return builder
}

// SetDocs sets the requested value.
func (builder *MCPServerBuilder) SetDocs(docs string) *MCPServerBuilder {
	builder.profile.MCPServer.Docs = strings.TrimSpace(docs)
	return builder
}

// SetVerification sets the requested value.
func (builder *MCPServerBuilder) SetVerification(verification MCPServerVerification) *MCPServerBuilder {
	builder.profile.MCPServer.Verification = &verification
	return builder
}

// AddVerificationDNS adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddVerificationDNS(domain string, dnsField string) *MCPServerBuilder {
	builder.profile.MCPServer.Verification = &MCPServerVerification{
		Type:     VerificationTypeDNS,
		Value:    strings.TrimSpace(domain),
		DNSField: strings.TrimSpace(dnsField),
	}
	return builder
}

// AddVerificationSignature adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddVerificationSignature(signature string) *MCPServerBuilder {
	builder.profile.MCPServer.Verification = &MCPServerVerification{
		Type:  VerificationTypeSignature,
		Value: strings.TrimSpace(signature),
	}
	return builder
}

// AddVerificationChallenge adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddVerificationChallenge(challengePath string) *MCPServerBuilder {
	builder.profile.MCPServer.Verification = &MCPServerVerification{
		Type:          VerificationTypeChallenge,
		ChallengePath: strings.TrimSpace(challengePath),
	}
	return builder
}

// AddSocial adds the provided value to the current configuration.
func (builder *MCPServerBuilder) AddSocial(platform string, handle string) *MCPServerBuilder {
	trimmedPlatform := strings.TrimSpace(platform)
	trimmedHandle := strings.TrimSpace(handle)
	if trimmedPlatform == "" || trimmedHandle == "" {
		return builder
	}
	for index := range builder.profile.Socials {
		if builder.profile.Socials[index].Platform == SocialPlatform(trimmedPlatform) {
			builder.profile.Socials[index].Handle = trimmedHandle
			return builder
		}
	}
	builder.profile.Socials = append(builder.profile.Socials, SocialLink{
		Platform: SocialPlatform(trimmedPlatform),
		Handle:   trimmedHandle,
	})
	return builder
}

// SetSocials sets the requested value.
func (builder *MCPServerBuilder) SetSocials(socials []SocialLink) *MCPServerBuilder {
	builder.profile.Socials = append([]SocialLink{}, socials...)
	return builder
}

// Build builds and returns the configured value.
func (builder *MCPServerBuilder) Build() (HCS11Profile, error) {
	if strings.TrimSpace(builder.profile.DisplayName) == "" {
		return HCS11Profile{}, fmt.Errorf("MCP server name is required")
	}
	if builder.profile.MCPServer == nil {
		return HCS11Profile{}, fmt.Errorf("MCP server details are required")
	}
	if strings.TrimSpace(builder.profile.MCPServer.Version) == "" {
		return HCS11Profile{}, fmt.Errorf("MCP server version is required")
	}
	if strings.TrimSpace(builder.profile.MCPServer.ConnectionInfo.URL) == "" {
		return HCS11Profile{}, fmt.Errorf("MCP server connection URL is required")
	}
	if strings.TrimSpace(builder.profile.MCPServer.ConnectionInfo.Transport) == "" {
		return HCS11Profile{}, fmt.Errorf("MCP server transport is required")
	}
	if len(builder.profile.MCPServer.Services) == 0 {
		return HCS11Profile{}, fmt.Errorf("at least one MCP service capability is required")
	}
	if strings.TrimSpace(builder.profile.MCPServer.Description) == "" {
		return HCS11Profile{}, fmt.Errorf("MCP server description is required")
	}
	return builder.profile, nil
}
