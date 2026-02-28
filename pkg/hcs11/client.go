package hcs11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashgraph-online/go-sdk/pkg/hcs14"
	"github.com/hashgraph-online/go-sdk/pkg/mirror"
	"github.com/hashgraph-online/go-sdk/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient       *hedera.Client
	mirrorClient       *mirror.Client
	operatorAccountID  string
	operatorPrivateKey string
	network            string
	keyType            string
	kiloScribeBaseURL  string
	inscriberAuthURL   string
	inscriberAPIURL    string
}

func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}
	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}

	operatorAccountID := strings.TrimSpace(config.Auth.OperatorID)
	operatorPrivateKey := strings.TrimSpace(config.Auth.PrivateKey)
	if operatorAccountID != "" && operatorPrivateKey != "" {
		accountID, parseErr := hedera.AccountIDFromString(operatorAccountID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid operator account ID: %w", parseErr)
		}
		privateKey, keyErr := shared.ParsePrivateKey(operatorPrivateKey)
		if keyErr != nil {
			return nil, keyErr
		}
		hederaClient.SetOperator(accountID, privateKey)
	}

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
	})
	if err != nil {
		return nil, err
	}

	keyType := strings.TrimSpace(config.KeyType)
	if keyType == "" {
		keyType = "ed25519"
	}

	kiloScribeBaseURL := strings.TrimSpace(config.KiloScribeBaseURL)
	if kiloScribeBaseURL == "" {
		kiloScribeBaseURL = "https://kiloscribe.com"
	}

	return &Client{
		hederaClient:       hederaClient,
		mirrorClient:       mirrorClient,
		operatorAccountID:  operatorAccountID,
		operatorPrivateKey: operatorPrivateKey,
		network:            network,
		keyType:            keyType,
		kiloScribeBaseURL:  strings.TrimRight(kiloScribeBaseURL, "/"),
		inscriberAuthURL:   strings.TrimSpace(config.InscriberAuthURL),
		inscriberAPIURL:    strings.TrimSpace(config.InscriberAPIURL),
	}, nil
}

func (c *Client) HederaClient() *hedera.Client {
	return c.hederaClient
}

func (c *Client) OperatorID() string {
	return c.operatorAccountID
}

func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

func (c *Client) CreatePersonalProfile(
	displayName string,
	options map[string]any,
) HCS11Profile {
	profile := HCS11Profile{
		Version:     "1.0",
		Type:        ProfileTypePersonal,
		DisplayName: strings.TrimSpace(displayName),
	}
	assignProfileOptions(&profile, options)
	return profile
}

func (c *Client) CreateAIAgentProfile(
	displayName string,
	agentType AIAgentType,
	capabilities []AIAgentCapability,
	model string,
	options map[string]any,
) (HCS11Profile, error) {
	profile := HCS11Profile{
		Version:     "1.0",
		Type:        ProfileTypeAIAgent,
		DisplayName: strings.TrimSpace(displayName),
		AIAgent: &AIAgentDetails{
			Type:         agentType,
			Capabilities: append([]AIAgentCapability{}, capabilities...),
			Model:        strings.TrimSpace(model),
		},
	}
	assignProfileOptions(&profile, options)
	validation := c.ValidateProfile(profile)
	if !validation.Valid {
		return HCS11Profile{}, fmt.Errorf("invalid AI Agent Profile: %s", strings.Join(validation.Errors, ", "))
	}
	return profile, nil
}

func (c *Client) CreateMCPServerProfile(
	displayName string,
	serverDetails MCPServerDetails,
	options map[string]any,
) (HCS11Profile, error) {
	profile := HCS11Profile{
		Version:     "1.0",
		Type:        ProfileTypeMCPServer,
		DisplayName: strings.TrimSpace(displayName),
		MCPServer:   &serverDetails,
	}
	assignProfileOptions(&profile, options)
	validation := c.ValidateProfile(profile)
	if !validation.Valid {
		return HCS11Profile{}, fmt.Errorf("invalid MCP Server Profile: %s", strings.Join(validation.Errors, ", "))
	}
	return profile, nil
}

func (c *Client) ValidateProfile(profile HCS11Profile) ValidationResult {
	errors := make([]string, 0)
	if strings.TrimSpace(profile.Version) == "" {
		errors = append(errors, "version is required")
	}
	if strings.TrimSpace(profile.DisplayName) == "" {
		errors = append(errors, "display_name is required")
	}

	switch profile.Type {
	case ProfileTypePersonal:
	case ProfileTypeAIAgent:
		if profile.AIAgent == nil {
			errors = append(errors, "aiAgent is required for AI_AGENT profile")
		} else {
			if len(profile.AIAgent.Capabilities) == 0 {
				errors = append(errors, "aiAgent.capabilities must not be empty")
			}
			if strings.TrimSpace(profile.AIAgent.Model) == "" {
				errors = append(errors, "aiAgent.model is required")
			}
		}
	case ProfileTypeMCPServer:
		if profile.MCPServer == nil {
			errors = append(errors, "mcpServer is required for MCP_SERVER profile")
		} else {
			if strings.TrimSpace(profile.MCPServer.Version) == "" {
				errors = append(errors, "mcpServer.version is required")
			}
			if strings.TrimSpace(profile.MCPServer.ConnectionInfo.URL) == "" {
				errors = append(errors, "mcpServer.connectionInfo.url is required")
			}
			transport := strings.TrimSpace(profile.MCPServer.ConnectionInfo.Transport)
			if transport == "" {
				errors = append(errors, "mcpServer.connectionInfo.transport is required")
			} else if transport != "stdio" && transport != "sse" {
				errors = append(errors, "mcpServer.connectionInfo.transport must be stdio or sse")
			}
			if len(profile.MCPServer.Services) == 0 {
				errors = append(errors, "mcpServer.services must not be empty")
			}
			if strings.TrimSpace(profile.MCPServer.Description) == "" {
				errors = append(errors, "mcpServer.description is required")
			}
			if profile.MCPServer.Verification != nil {
				switch profile.MCPServer.Verification.Type {
				case VerificationTypeDNS, VerificationTypeSignature, VerificationTypeChallenge:
				default:
					errors = append(errors, "mcpServer.verification.type is invalid")
				}
			}
		}
	case ProfileTypeFlora:
		if len(profile.Members) == 0 {
			errors = append(errors, "flora members are required")
		}
		if profile.Threshold < 1 {
			errors = append(errors, "flora threshold must be at least 1")
		}
		if profile.Topics == nil {
			errors = append(errors, "flora topics are required")
		}
	default:
		errors = append(errors, "profile type is invalid")
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (c *Client) ProfileToJSONString(profile HCS11Profile) (string, error) {
	encodedProfile, err := json.Marshal(profile)
	if err != nil {
		return "", err
	}
	return string(encodedProfile), nil
}

func (c *Client) ParseProfileFromString(profileString string) (*HCS11Profile, error) {
	var profile HCS11Profile
	if err := json.Unmarshal([]byte(profileString), &profile); err != nil {
		return nil, err
	}
	validation := c.ValidateProfile(profile)
	if !validation.Valid {
		return nil, fmt.Errorf("invalid profile format: %s", strings.Join(validation.Errors, ", "))
	}
	return &profile, nil
}

func (c *Client) SetProfileForAccountMemo(topicID string, topicStandard int) string {
	if topicStandard == 0 {
		topicStandard = 1
	}
	return fmt.Sprintf("hcs-11:hcs://%d/%s", topicStandard, strings.TrimSpace(topicID))
}

func (c *Client) GetCapabilitiesFromTags(capabilityNames []string) []int {
	if len(capabilityNames) == 0 {
		return []int{int(AIAgentCapabilityTextGeneration)}
	}
	capabilities := make([]int, 0)
	for _, capabilityName := range capabilityNames {
		capability, ok := CapabilityNameToCapabilityMap[strings.ToLower(strings.TrimSpace(capabilityName))]
		if ok && !containsInt(capabilities, int(capability)) {
			capabilities = append(capabilities, int(capability))
		}
	}
	if len(capabilities) == 0 {
		return []int{int(AIAgentCapabilityTextGeneration)}
	}
	return capabilities
}

func (c *Client) GetAgentTypeFromMetadata(metadata AgentMetadata) AIAgentType {
	if strings.EqualFold(metadata.Type, "autonomous") {
		return AIAgentTypeAutonomous
	}
	return AIAgentTypeManual
}

func (c *Client) AttachUAIDIfMissing(_ context.Context, profile *HCS11Profile) error {
	if profile == nil || strings.TrimSpace(profile.UAID) != "" {
		return nil
	}
	if strings.TrimSpace(c.operatorAccountID) == "" {
		return nil
	}

	nativeID := fmt.Sprintf("hedera:%s:%s", c.network, c.operatorAccountID)
	uid := c.operatorAccountID
	if strings.TrimSpace(profile.InboundTopicID) != "" {
		uid = fmt.Sprintf("%s@%s", profile.InboundTopicID, c.operatorAccountID)
	}
	uaid, err := hcs14.CreateUAIDFromDID(
		fmt.Sprintf("did:hedera:%s:%s", c.network, c.operatorAccountID),
		hcs14.RoutingParams{
			UID:      uid,
			Proto:    "hcs-10",
			NativeID: nativeID,
		},
	)
	if err != nil {
		return err
	}
	profile.UAID = uaid
	return nil
}

func containsInt(input []int, target int) bool {
	for _, value := range input {
		if value == target {
			return true
		}
	}
	return false
}

func assignProfileOptions(profile *HCS11Profile, options map[string]any) {
	if profile == nil || options == nil {
		return
	}
	if alias, ok := options["alias"].(string); ok {
		profile.Alias = strings.TrimSpace(alias)
	}
	if bio, ok := options["bio"].(string); ok {
		profile.Bio = strings.TrimSpace(bio)
	}
	if profileImage, ok := options["profileImage"].(string); ok {
		profile.ProfileImage = strings.TrimSpace(profileImage)
	}
	if inboundTopicID, ok := options["inboundTopicId"].(string); ok {
		profile.InboundTopicID = strings.TrimSpace(inboundTopicID)
	}
	if outboundTopicID, ok := options["outboundTopicId"].(string); ok {
		profile.OutboundTopicID = strings.TrimSpace(outboundTopicID)
	}
	if baseAccount, ok := options["baseAccount"].(string); ok {
		profile.BaseAccount = strings.TrimSpace(baseAccount)
	}
}
