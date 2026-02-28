package hcs11

import (
	"fmt"
	"strings"
)

type PersonBuilder struct {
	profile HCS11Profile
}

func NewPersonBuilder() *PersonBuilder {
	return &PersonBuilder{
		profile: HCS11Profile{
			Version: "1.0",
			Type:    ProfileTypePersonal,
		},
	}
}

func (builder *PersonBuilder) SetName(name string) *PersonBuilder {
	builder.profile.DisplayName = strings.TrimSpace(name)
	return builder
}

func (builder *PersonBuilder) SetAlias(alias string) *PersonBuilder {
	builder.profile.Alias = strings.TrimSpace(alias)
	return builder
}

func (builder *PersonBuilder) SetBio(bio string) *PersonBuilder {
	builder.profile.Bio = strings.TrimSpace(bio)
	return builder
}

func (builder *PersonBuilder) SetDescription(description string) *PersonBuilder {
	return builder.SetBio(description)
}

func (builder *PersonBuilder) AddSocial(platform string, handle string) *PersonBuilder {
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

func (builder *PersonBuilder) SetProfileImage(profileImage string) *PersonBuilder {
	builder.profile.ProfileImage = strings.TrimSpace(profileImage)
	return builder
}

func (builder *PersonBuilder) SetBaseAccount(accountID string) *PersonBuilder {
	builder.profile.BaseAccount = strings.TrimSpace(accountID)
	return builder
}

func (builder *PersonBuilder) SetInboundTopicID(topicID string) *PersonBuilder {
	builder.profile.InboundTopicID = strings.TrimSpace(topicID)
	return builder
}

func (builder *PersonBuilder) SetOutboundTopicID(topicID string) *PersonBuilder {
	builder.profile.OutboundTopicID = strings.TrimSpace(topicID)
	return builder
}

func (builder *PersonBuilder) AddProperty(key string, value any) *PersonBuilder {
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

func (builder *PersonBuilder) Build() (HCS11Profile, error) {
	if strings.TrimSpace(builder.profile.DisplayName) == "" {
		return HCS11Profile{}, fmt.Errorf("display name is required for person profile")
	}
	return builder.profile, nil
}
