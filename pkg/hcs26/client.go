package hcs26

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

type Client struct {
	network      string
	mirrorClient *mirror.Client
}

// NewClient creates a new Client.
func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}
	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
		APIKey:  config.MirrorAPIKey,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		network:      network,
		mirrorClient: mirrorClient,
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// ResolveDiscoveryRecord performs the requested operation.
func (c *Client) ResolveDiscoveryRecord(
	ctx context.Context,
	directoryTopicID string,
	skillUID int64,
	scanLimit int,
) (*DiscoveryRegister, error) {
	register, err := c.getDiscoveryRegister(ctx, directoryTopicID, skillUID)
	if err != nil || register == nil {
		return register, err
	}

	limit := scanLimit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 5000 {
		limit = 5000
	}

	items, err := c.mirrorClient.GetTopicMessages(ctx, directoryTopicID, mirror.MessageQueryOptions{
		Limit: limit,
		Order: "asc",
	})
	if err != nil {
		return nil, err
	}

	registerSequence := register.SequenceNumber
	uid := fmt.Sprintf("%d", skillUID)
	current := *register
	for _, item := range items {
		payload, err := decodePayload(item.Message)
		if err != nil {
			continue
		}
		p, _ := payload["p"].(string)
		op, _ := payload["op"].(string)
		if strings.TrimSpace(p) != Protocol {
			continue
		}
		if op == "delete" {
			deleteUID, _ := payload["uid"].(string)
			if strings.TrimSpace(deleteUID) != uid {
				continue
			}
			deleteSequence := sequenceFromPayload(payload, item.SequenceNumber)
			if deleteSequence > registerSequence {
				return nil, nil
			}
		}
		if op == "update" {
			updateUID, _ := payload["uid"].(string)
			if strings.TrimSpace(updateUID) != uid {
				continue
			}
			updateSequence := sequenceFromPayload(payload, item.SequenceNumber)
			if updateSequence <= registerSequence {
				continue
			}
			metadataValue, hasMetadata := payload["metadata"]
			if hasMetadata {
				resolvedMetadata, err := c.resolveDiscoveryMetadata(ctx, metadataValue, true)
				if err != nil {
					return nil, err
				}
				if currentMetadata, ok := current.Metadata.(map[string]any); ok {
					for key, value := range resolvedMetadata {
						currentMetadata[key] = value
					}
					current.Metadata = currentMetadata
				} else {
					current.Metadata = resolvedMetadata
				}
			}
			if accountID, hasAccountID := payload["account_id"].(string); hasAccountID && strings.TrimSpace(accountID) != "" {
				current.AccountID = strings.TrimSpace(accountID)
			}
		}
	}
	return &current, nil
}

func (c *Client) getDiscoveryRegister(ctx context.Context, directoryTopicID string, skillUID int64) (*DiscoveryRegister, error) {
	items, err := c.mirrorClient.GetTopicMessages(ctx, directoryTopicID, mirror.MessageQueryOptions{
		SequenceNumber: fmt.Sprintf("eq:%d", skillUID),
		Limit:          5,
		Order:          "asc",
	})
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		payload, err := decodePayload(item.Message)
		if err != nil {
			continue
		}
		register, ok, err := parseDiscoveryRegister(payload, item.SequenceNumber)
		if err != nil || !ok {
			continue
		}
		if register.SequenceNumber != skillUID {
			continue
		}
		resolvedMetadata, err := c.resolveDiscoveryMetadata(ctx, register.Metadata, false)
		if err != nil {
			return nil, err
		}
		register.Metadata = resolvedMetadata
		return &register, nil
	}

	return nil, nil
}

func (c *Client) resolveDiscoveryMetadata(ctx context.Context, metadataValue any, patch bool) (map[string]any, error) {
	switch typed := metadataValue.(type) {
	case string:
		if !strings.HasPrefix(strings.TrimSpace(typed), "hcs://1/") {
			return nil, fmt.Errorf("unsupported discovery metadata URI: %s", typed)
		}
		topicID := strings.TrimPrefix(strings.TrimSpace(typed), "hcs://1/")
		topicID = strings.Split(topicID, "/")[0]
		rawJSON, err := c.fetchTopicMessageJSON(ctx, topicID)
		if err != nil {
			return nil, err
		}
		var parsed map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
			return nil, fmt.Errorf("discovery metadata is not valid JSON: %w", err)
		}
		if !patch {
			if err := validateDiscoveryMetadata(parsed); err != nil {
				return nil, err
			}
		}
		return normalizeDiscoveryMetadata(parsed), nil
	case map[string]any:
		if !patch {
			if err := validateDiscoveryMetadata(typed); err != nil {
				return nil, err
			}
		}
		return normalizeDiscoveryMetadata(typed), nil
	default:
		return nil, fmt.Errorf("unsupported discovery metadata type")
	}
}

func normalizeDiscoveryMetadata(metadata map[string]any) map[string]any {
	iconValue, hasIcon := metadata["icon"].(string)
	if !hasIcon || strings.TrimSpace(iconValue) == "" {
		if legacy, ok := metadata["icon_hcs1"].(string); ok && strings.TrimSpace(legacy) != "" {
			metadata["icon"] = strings.TrimSpace(legacy)
		}
	}
	return metadata
}

// ListVersionRegisters performs the requested operation.
func (c *Client) ListVersionRegisters(
	ctx context.Context,
	versionRegistryTopicID string,
	skillUID int64,
	limit int,
) ([]any, error) {
	maxLimit := limit
	if maxLimit <= 0 {
		maxLimit = 250
	}
	if maxLimit > 1000 {
		maxLimit = 1000
	}
	items, err := c.mirrorClient.GetTopicMessages(ctx, versionRegistryTopicID, mirror.MessageQueryOptions{
		Limit: maxLimit,
		Order: "asc",
	})
	if err != nil {
		return nil, err
	}

	statusByUID := map[string]string{}
	result := make([]any, 0)
	for _, item := range items {
		payload, err := decodePayload(item.Message)
		if err != nil {
			continue
		}
		p, _ := payload["p"].(string)
		op, _ := payload["op"].(string)
		if strings.TrimSpace(p) != Protocol {
			continue
		}
		switch op {
		case "update":
			uid, _ := payload["uid"].(string)
			status, _ := payload["status"].(string)
			if strings.TrimSpace(uid) != "" && strings.TrimSpace(status) != "" {
				statusByUID[strings.TrimSpace(uid)] = strings.TrimSpace(status)
			}
		case "delete":
			uid, _ := payload["uid"].(string)
			if strings.TrimSpace(uid) != "" {
				statusByUID[strings.TrimSpace(uid)] = "yanked"
			}
		case "register":
			entry, ok, err := parseVersionRegister(payload, item.SequenceNumber)
			if err != nil || !ok {
				continue
			}
			uidString := fmt.Sprintf("%d", item.SequenceNumber)
			if overrideStatus, hasStatus := statusByUID[uidString]; hasStatus {
				switch typed := entry.(type) {
				case VersionRegister:
					typed.Status = overrideStatus
					entry = typed
				case VersionRegisterLegacy:
					typed.Status = overrideStatus
					entry = typed
				}
			}
			var entrySkillUID int64
			var entryStatus string
			switch typed := entry.(type) {
			case VersionRegister:
				entrySkillUID = typed.SkillUID
				entryStatus = typed.Status
			case VersionRegisterLegacy:
				entrySkillUID = typed.SkillUID
				entryStatus = typed.Status
			}
			if entrySkillUID == skillUID && (entryStatus == "" || entryStatus == "active") {
				result = append(result, entry)
			}
		}
	}
	return result, nil
}

// GetLatestVersionRegister performs the requested operation.
func (c *Client) GetLatestVersionRegister(
	ctx context.Context,
	versionRegistryTopicID string,
	skillUID int64,
) (any, error) {
	entries, err := c.ListVersionRegisters(ctx, versionRegistryTopicID, skillUID, 250)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	sort.SliceStable(entries, func(i int, j int) bool {
		leftVersion := extractVersion(entries[i])
		rightVersion := extractVersion(entries[j])
		return compareSemver(leftVersion, rightVersion) < 0
	})
	return entries[len(entries)-1], nil
}

// ResolveManifest performs the requested operation.
func (c *Client) ResolveManifest(ctx context.Context, manifestTopicID string) (SkillManifest, string, error) {
	rawJSON, err := c.fetchTopicMessageJSON(ctx, manifestTopicID)
	if err != nil {
		return SkillManifest{}, "", err
	}

	var manifest SkillManifest
	if err := json.Unmarshal([]byte(rawJSON), &manifest); err != nil {
		return SkillManifest{}, "", fmt.Errorf("manifest content is not valid JSON: %w", err)
	}
	if err := validateManifest(manifest); err != nil {
		return SkillManifest{}, "", err
	}

	sum := sha256.Sum256([]byte(rawJSON))
	return manifest, hex.EncodeToString(sum[:]), nil
}

// VerifyVersionRegisterMatchesManifest performs the requested operation.
func (c *Client) VerifyVersionRegisterMatchesManifest(versionRegister any, manifestSHA256Hex string) error {
	expected := strings.ToLower(strings.TrimSpace(manifestSHA256Hex))
	switch typed := versionRegister.(type) {
	case VersionRegister:
		if strings.TrimSpace(typed.Checksum) == "" {
			return nil
		}
		if !checksumPattern.MatchString(strings.ToLower(strings.TrimSpace(typed.Checksum))) {
			return fmt.Errorf("invalid checksum format on version register")
		}
		if strings.TrimPrefix(strings.ToLower(strings.TrimSpace(typed.Checksum)), "sha256:") != expected {
			return fmt.Errorf("version register checksum does not match manifest SHA-256")
		}
	case VersionRegisterLegacy:
		if strings.TrimSpace(typed.Checksum) == "" {
			return nil
		}
		if !checksumPattern.MatchString(strings.ToLower(strings.TrimSpace(typed.Checksum))) {
			return fmt.Errorf("invalid checksum format on version register")
		}
		if strings.TrimPrefix(strings.ToLower(strings.TrimSpace(typed.Checksum)), "sha256:") != expected {
			return fmt.Errorf("version register checksum does not match manifest SHA-256")
		}
	default:
		return fmt.Errorf("unsupported version register type")
	}
	return nil
}

// ResolveSkill performs the requested operation.
func (c *Client) ResolveSkill(
	ctx context.Context,
	directoryTopicID string,
	skillUID int64,
	discoveryScanLimit int,
) (*ResolvedSkill, error) {
	discovery, err := c.ResolveDiscoveryRecord(ctx, directoryTopicID, skillUID, discoveryScanLimit)
	if err != nil || discovery == nil {
		return nil, err
	}

	latestVersion, err := c.GetLatestVersionRegister(ctx, discovery.VersionRegistry, skillUID)
	if err != nil || latestVersion == nil {
		return nil, err
	}

	manifestTopicID, err := manifestTopicIDFromVersion(latestVersion)
	if err != nil {
		return nil, err
	}
	manifest, manifestHash, err := c.ResolveManifest(ctx, manifestTopicID)
	if err != nil {
		return nil, err
	}
	if err := c.VerifyVersionRegisterMatchesManifest(latestVersion, manifestHash); err != nil {
		return nil, err
	}

	return &ResolvedSkill{
		DirectoryTopicID:       directoryTopicID,
		SkillUID:               skillUID,
		Discovery:              *discovery,
		VersionRegistryTopicID: discovery.VersionRegistry,
		LatestVersion:          latestVersion,
		Manifest:               manifest,
		ManifestSHA256Hex:      manifestHash,
	}, nil
}

func (c *Client) fetchTopicMessageJSON(ctx context.Context, topicID string) (string, error) {
	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Limit: 1,
		Order: "desc",
	})
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no messages found in topic %s", topicID)
	}
	decoded, err := base64.StdEncoding.DecodeString(items[0].Message)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func parseDiscoveryRegister(payload map[string]any, fallbackSequence int64) (DiscoveryRegister, bool, error) {
	protocol, _ := payload["p"].(string)
	operation, _ := payload["op"].(string)
	if strings.TrimSpace(protocol) != Protocol || strings.TrimSpace(operation) != "register" {
		return DiscoveryRegister{}, false, nil
	}

	versionRegistry, hasVersionRegistry := payload["t_id"].(string)
	accountID, hasAccountID := payload["account_id"].(string)
	if hasVersionRegistry && hasAccountID {
		register := DiscoveryRegister{
			P:               Protocol,
			Op:              "register",
			VersionRegistry: strings.TrimSpace(versionRegistry),
			AccountID:       strings.TrimSpace(accountID),
			Metadata:        payload["metadata"],
			Memo:            readString(payload, "m"),
			SequenceNumber:  sequenceFromPayload(payload, fallbackSequence),
		}
		return register, true, nil
	}

	legacyVersionRegistry, hasLegacyVersionRegistry := payload["version_registry"].(string)
	legacyPublisher, hasLegacyPublisher := payload["publisher"].(string)
	if !hasLegacyVersionRegistry || !hasLegacyPublisher {
		return DiscoveryRegister{}, false, nil
	}
	register := DiscoveryRegister{
		P:               Protocol,
		Op:              "register",
		VersionRegistry: strings.TrimSpace(legacyVersionRegistry),
		AccountID:       strings.TrimSpace(legacyPublisher),
		Metadata:        payload["metadata"],
		Memo:            readString(payload, "m"),
		SequenceNumber:  sequenceFromPayload(payload, fallbackSequence),
	}
	return register, true, nil
}

func parseVersionRegister(payload map[string]any, fallbackSequence int64) (any, bool, error) {
	protocol, _ := payload["p"].(string)
	operation, _ := payload["op"].(string)
	if strings.TrimSpace(protocol) != Protocol || strings.TrimSpace(operation) != "register" {
		return nil, false, nil
	}

	skillUID, hasSkillUID := toInt64(payload["skill_uid"])
	version, hasVersion := payload["version"].(string)
	if !hasSkillUID || !hasVersion || !semverPattern.MatchString(strings.TrimSpace(version)) {
		return nil, false, nil
	}

	if manifestTopicID, hasManifestTopicID := payload["t_id"].(string); hasManifestTopicID {
		entry := VersionRegister{
			P:               Protocol,
			Op:              "register",
			SkillUID:        skillUID,
			Version:         strings.TrimSpace(version),
			ManifestTopicID: strings.TrimSpace(manifestTopicID),
			Checksum:        readString(payload, "checksum"),
			Status:          readString(payload, "status"),
			Memo:            readString(payload, "m"),
			SequenceNumber:  sequenceFromPayload(payload, fallbackSequence),
		}
		return entry, true, nil
	}

	manifestHRL, hasManifestHRL := payload["manifest_hcs1"].(string)
	if !hasManifestHRL {
		return nil, false, nil
	}
	entry := VersionRegisterLegacy{
		P:              Protocol,
		Op:             "register",
		SkillUID:       skillUID,
		Version:        strings.TrimSpace(version),
		ManifestHRL:    strings.TrimSpace(manifestHRL),
		Checksum:       readString(payload, "checksum"),
		Status:         readString(payload, "status"),
		Memo:           readString(payload, "m"),
		SequenceNumber: sequenceFromPayload(payload, fallbackSequence),
	}
	return entry, true, nil
}

func decodePayload(encoded string) (map[string]any, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func sequenceFromPayload(payload map[string]any, fallback int64) int64 {
	if value, ok := toInt64(payload["sequence_number"]); ok {
		return value
	}
	return fallback
}

func toInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func readString(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(typed)
}

func validateDiscoveryMetadata(metadata map[string]any) error {
	name, hasName := metadata["name"].(string)
	description, hasDescription := metadata["description"].(string)
	license, hasLicense := metadata["license"].(string)
	if !hasName || strings.TrimSpace(name) == "" || !hasDescription || strings.TrimSpace(description) == "" ||
		!hasLicense || strings.TrimSpace(license) == "" {
		return fmt.Errorf("discovery metadata must include name, description, and license")
	}
	if _, hasAuthor := metadata["author"]; !hasAuthor {
		return fmt.Errorf("discovery metadata must include author")
	}
	return nil
}

func validateManifest(manifest SkillManifest) error {
	if strings.TrimSpace(manifest.Name) == "" || strings.TrimSpace(manifest.Description) == "" ||
		strings.TrimSpace(manifest.Version) == "" || strings.TrimSpace(manifest.License) == "" {
		return fmt.Errorf("manifest name, description, version, and license are required")
	}
	if !semverPattern.MatchString(strings.TrimSpace(manifest.Version)) {
		return fmt.Errorf("manifest version must be semantic version")
	}
	if len(manifest.Files) == 0 {
		return fmt.Errorf("manifest files list cannot be empty")
	}
	hasSkillMD := false
	for _, file := range manifest.Files {
		if strings.TrimSpace(file.Path) == "SKILL.md" {
			hasSkillMD = true
		}
		if strings.TrimSpace(file.Path) == "" || strings.TrimSpace(file.HRL) == "" ||
			strings.TrimSpace(file.SHA256) == "" || strings.TrimSpace(file.Mime) == "" {
			return fmt.Errorf("manifest file entries require path, hrl, sha256, mime")
		}
		if !hrlPattern.MatchString(strings.TrimSpace(file.HRL)) {
			return fmt.Errorf("manifest file hrl must be hcs://1/<topicId>")
		}
		if !regexpSHA256(file.SHA256) {
			return fmt.Errorf("manifest file sha256 must be 64 hex characters")
		}
	}
	if !hasSkillMD {
		return fmt.Errorf("manifest files must include SKILL.md")
	}
	return nil
}

func regexpSHA256(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if len(trimmed) != 64 {
		return false
	}
	for _, character := range trimmed {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func manifestTopicIDFromVersion(versionEntry any) (string, error) {
	switch typed := versionEntry.(type) {
	case VersionRegister:
		return strings.TrimSpace(typed.ManifestTopicID), nil
	case VersionRegisterLegacy:
		hrl := strings.TrimSpace(typed.ManifestHRL)
		if !strings.HasPrefix(hrl, "hcs://1/") {
			return "", fmt.Errorf("invalid manifest HRL: %s", hrl)
		}
		return strings.TrimPrefix(hrl, "hcs://1/"), nil
	default:
		return "", fmt.Errorf("unsupported version entry type")
	}
}

type parsedSemver struct {
	major int64
	minor int64
	patch int64
}

func extractVersion(entry any) parsedSemver {
	versionRaw := "0.0.0"
	switch typed := entry.(type) {
	case VersionRegister:
		versionRaw = typed.Version
	case VersionRegisterLegacy:
		versionRaw = typed.Version
	}
	return parseSemver(versionRaw)
}

func parseSemver(versionRaw string) parsedSemver {
	trimmed := strings.TrimPrefix(strings.TrimSpace(versionRaw), "v")
	matches := semverPattern.FindStringSubmatch(trimmed)
	if len(matches) < 4 {
		return parsedSemver{}
	}
	var major, minor, patch int64
	_, _ = fmt.Sscanf(matches[1], "%d", &major)
	_, _ = fmt.Sscanf(matches[2], "%d", &minor)
	_, _ = fmt.Sscanf(matches[3], "%d", &patch)
	return parsedSemver{major: major, minor: minor, patch: patch}
}

func compareSemver(left parsedSemver, right parsedSemver) int {
	if left.major != right.major {
		if left.major < right.major {
			return -1
		}
		return 1
	}
	if left.minor != right.minor {
		if left.minor < right.minor {
			return -1
		}
		return 1
	}
	if left.patch != right.patch {
		if left.patch < right.patch {
			return -1
		}
		return 1
	}
	return 0
}
