package hcs11

import (
	"fmt"
	"strings"
)

type FloraBuilder struct {
	profile HCS11Profile
}

func NewFloraBuilder() *FloraBuilder {
	return &FloraBuilder{
		profile: HCS11Profile{
			Version: "1.0",
			Type:    ProfileTypeFlora,
		},
	}
}

func (builder *FloraBuilder) SetDisplayName(displayName string) *FloraBuilder {
	builder.profile.DisplayName = strings.TrimSpace(displayName)
	return builder
}

func (builder *FloraBuilder) SetBio(bio string) *FloraBuilder {
	builder.profile.Bio = strings.TrimSpace(bio)
	return builder
}

func (builder *FloraBuilder) SetMembers(members []FloraMember) *FloraBuilder {
	builder.profile.Members = append([]FloraMember{}, members...)
	return builder
}

func (builder *FloraBuilder) SetThreshold(threshold int) *FloraBuilder {
	builder.profile.Threshold = threshold
	return builder
}

func (builder *FloraBuilder) SetTopics(topics FloraTopics) *FloraBuilder {
	builder.profile.Topics = &FloraTopics{
		Communication: strings.TrimSpace(topics.Communication),
		Transaction:   strings.TrimSpace(topics.Transaction),
		State:         strings.TrimSpace(topics.State),
	}
	builder.profile.InboundTopicID = builder.profile.Topics.Communication
	builder.profile.OutboundTopicID = builder.profile.Topics.Transaction
	return builder
}

func (builder *FloraBuilder) SetPolicies(policies map[string]any) *FloraBuilder {
	builder.profile.Policies = copyStringAnyMap(policies)
	return builder
}

func (builder *FloraBuilder) SetMetadata(metadata map[string]any) *FloraBuilder {
	builder.profile.Metadata = copyStringAnyMap(metadata)
	return builder
}

func (builder *FloraBuilder) AddMetadata(key string, value any) *FloraBuilder {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return builder
	}
	if builder.profile.Metadata == nil {
		builder.profile.Metadata = map[string]any{}
	}
	builder.profile.Metadata[trimmedKey] = value
	return builder
}

func (builder *FloraBuilder) Build() (HCS11Profile, error) {
	if strings.TrimSpace(builder.profile.DisplayName) == "" {
		return HCS11Profile{}, fmt.Errorf("flora display name is required")
	}
	if len(builder.profile.Members) == 0 {
		return HCS11Profile{}, fmt.Errorf("flora must have at least one member")
	}
	if builder.profile.Threshold < 1 {
		return HCS11Profile{}, fmt.Errorf("flora threshold must be at least 1")
	}
	if builder.profile.Threshold > len(builder.profile.Members) {
		return HCS11Profile{}, fmt.Errorf("flora threshold cannot exceed number of members")
	}
	if builder.profile.Topics == nil {
		return HCS11Profile{}, fmt.Errorf("flora topics are required")
	}
	if strings.TrimSpace(builder.profile.InboundTopicID) == "" ||
		strings.TrimSpace(builder.profile.OutboundTopicID) == "" {
		return HCS11Profile{}, fmt.Errorf("flora inbound and outbound topic IDs are required")
	}
	return builder.profile, nil
}

func copyStringAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
