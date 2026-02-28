package hcs11

import (
	"fmt"
	"strings"
)

type AgentBuilder struct {
	profile HCS11Profile
}

// NewAgentBuilder creates a new AgentBuilder.
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{
		profile: HCS11Profile{
			Version: "1.0",
			Type:    ProfileTypeAIAgent,
			AIAgent: &AIAgentDetails{
				Type:         AIAgentTypeManual,
				Capabilities: []AIAgentCapability{},
			},
		},
	}
}

// SetName sets the requested value.
func (builder *AgentBuilder) SetName(name string) *AgentBuilder {
	builder.profile.DisplayName = strings.TrimSpace(name)
	return builder
}

// SetAlias sets the requested value.
func (builder *AgentBuilder) SetAlias(alias string) *AgentBuilder {
	builder.profile.Alias = strings.TrimSpace(alias)
	return builder
}

// SetBio sets the requested value.
func (builder *AgentBuilder) SetBio(bio string) *AgentBuilder {
	builder.profile.Bio = strings.TrimSpace(bio)
	return builder
}

// SetDescription sets the requested value.
func (builder *AgentBuilder) SetDescription(description string) *AgentBuilder {
	return builder.SetBio(description)
}

// SetCapabilities sets the requested value.
func (builder *AgentBuilder) SetCapabilities(capabilities []AIAgentCapability) *AgentBuilder {
	builder.profile.AIAgent.Capabilities = append([]AIAgentCapability{}, capabilities...)
	return builder
}

// SetType sets the requested value.
func (builder *AgentBuilder) SetType(agentType AIAgentType) *AgentBuilder {
	builder.profile.AIAgent.Type = agentType
	return builder
}

// SetModel sets the requested value.
func (builder *AgentBuilder) SetModel(model string) *AgentBuilder {
	builder.profile.AIAgent.Model = strings.TrimSpace(model)
	return builder
}

// SetCreator sets the requested value.
func (builder *AgentBuilder) SetCreator(creator string) *AgentBuilder {
	builder.profile.AIAgent.Creator = strings.TrimSpace(creator)
	return builder
}

// AddSocial adds the provided value to the current configuration.
func (builder *AgentBuilder) AddSocial(platform string, handle string) *AgentBuilder {
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

// AddProperty adds the provided value to the current configuration.
func (builder *AgentBuilder) AddProperty(key string, value any) *AgentBuilder {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return builder
	}
	if builder.profile.Properties == nil {
		builder.profile.Properties = map[string]any{}
	}
	builder.profile.Properties[trimmedKey] = value
	return builder
}

// SetInboundTopicID sets the requested value.
func (builder *AgentBuilder) SetInboundTopicID(topicID string) *AgentBuilder {
	builder.profile.InboundTopicID = strings.TrimSpace(topicID)
	return builder
}

// SetOutboundTopicID sets the requested value.
func (builder *AgentBuilder) SetOutboundTopicID(topicID string) *AgentBuilder {
	builder.profile.OutboundTopicID = strings.TrimSpace(topicID)
	return builder
}

// SetBaseAccount sets the requested value.
func (builder *AgentBuilder) SetBaseAccount(accountID string) *AgentBuilder {
	builder.profile.BaseAccount = strings.TrimSpace(accountID)
	return builder
}

// Build builds and returns the configured value.
func (builder *AgentBuilder) Build() (HCS11Profile, error) {
	if strings.TrimSpace(builder.profile.DisplayName) == "" {
		return HCS11Profile{}, fmt.Errorf("agent display name is required")
	}
	if builder.profile.AIAgent == nil {
		return HCS11Profile{}, fmt.Errorf("agent details are required")
	}
	return builder.profile, nil
}
