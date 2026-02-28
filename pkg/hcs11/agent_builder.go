package hcs11

import (
	"fmt"
	"strings"
)

type AgentBuilder struct {
	profile HCS11Profile
}

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

func (builder *AgentBuilder) SetName(name string) *AgentBuilder {
	builder.profile.DisplayName = strings.TrimSpace(name)
	return builder
}

func (builder *AgentBuilder) SetAlias(alias string) *AgentBuilder {
	builder.profile.Alias = strings.TrimSpace(alias)
	return builder
}

func (builder *AgentBuilder) SetBio(bio string) *AgentBuilder {
	builder.profile.Bio = strings.TrimSpace(bio)
	return builder
}

func (builder *AgentBuilder) SetDescription(description string) *AgentBuilder {
	return builder.SetBio(description)
}

func (builder *AgentBuilder) SetCapabilities(capabilities []AIAgentCapability) *AgentBuilder {
	builder.profile.AIAgent.Capabilities = append([]AIAgentCapability{}, capabilities...)
	return builder
}

func (builder *AgentBuilder) SetType(agentType AIAgentType) *AgentBuilder {
	builder.profile.AIAgent.Type = agentType
	return builder
}

func (builder *AgentBuilder) SetModel(model string) *AgentBuilder {
	builder.profile.AIAgent.Model = strings.TrimSpace(model)
	return builder
}

func (builder *AgentBuilder) SetCreator(creator string) *AgentBuilder {
	builder.profile.AIAgent.Creator = strings.TrimSpace(creator)
	return builder
}

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

func (builder *AgentBuilder) SetInboundTopicID(topicID string) *AgentBuilder {
	builder.profile.InboundTopicID = strings.TrimSpace(topicID)
	return builder
}

func (builder *AgentBuilder) SetOutboundTopicID(topicID string) *AgentBuilder {
	builder.profile.OutboundTopicID = strings.TrimSpace(topicID)
	return builder
}

func (builder *AgentBuilder) SetBaseAccount(accountID string) *AgentBuilder {
	builder.profile.BaseAccount = strings.TrimSpace(accountID)
	return builder
}

func (builder *AgentBuilder) Build() (HCS11Profile, error) {
	if strings.TrimSpace(builder.profile.DisplayName) == "" {
		return HCS11Profile{}, fmt.Errorf("agent display name is required")
	}
	if builder.profile.AIAgent == nil {
		return HCS11Profile{}, fmt.Errorf("agent details are required")
	}
	return builder.profile, nil
}
