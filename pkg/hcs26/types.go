package hcs26

import "regexp"

const (
	Protocol                = "hcs-26"
	DefaultTTLSeconds int64 = 86400
)

type TopicType int

const (
	TopicTypeDiscovery  TopicType = 0
	TopicTypeVersion    TopicType = 1
	TopicTypeReputation TopicType = 2
)

type Operation int

const (
	OperationRegister Operation = 0
	OperationUpdate   Operation = 1
	OperationDelete   Operation = 2
	OperationMigrate  Operation = 3
)

type DiscoveryMetadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Author       any      `json:"author"`
	License      string   `json:"license"`
	Tags         []int64  `json:"tags,omitempty"`
	Homepage     string   `json:"homepage,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	Languages    []string `json:"languages,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Repo         string   `json:"repo,omitempty"`
	Commit       string   `json:"commit,omitempty"`
}

type DiscoveryRegister struct {
	P               string `json:"p"`
	Op              string `json:"op"`
	VersionRegistry string `json:"t_id"`
	AccountID       string `json:"account_id"`
	Metadata        any    `json:"metadata"`
	Memo            string `json:"m,omitempty"`
	SequenceNumber  int64  `json:"sequence_number,omitempty"`
}

type DiscoveryRegisterLegacy struct {
	P               string `json:"p"`
	Op              string `json:"op"`
	VersionRegistry string `json:"version_registry"`
	Publisher       string `json:"publisher"`
	Metadata        any    `json:"metadata"`
	Memo            string `json:"m,omitempty"`
	SequenceNumber  int64  `json:"sequence_number,omitempty"`
}

type DiscoveryUpdate struct {
	P              string `json:"p"`
	Op             string `json:"op"`
	UID            string `json:"uid"`
	AccountID      string `json:"account_id,omitempty"`
	Metadata       any    `json:"metadata,omitempty"`
	Memo           string `json:"m,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type DiscoveryDelete struct {
	P              string `json:"p"`
	Op             string `json:"op"`
	UID            string `json:"uid"`
	Memo           string `json:"m,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type VersionRegister struct {
	P               string `json:"p"`
	Op              string `json:"op"`
	SkillUID        int64  `json:"skill_uid"`
	Version         string `json:"version"`
	ManifestTopicID string `json:"t_id"`
	Checksum        string `json:"checksum,omitempty"`
	Status          string `json:"status,omitempty"`
	Memo            string `json:"m,omitempty"`
	SequenceNumber  int64  `json:"sequence_number,omitempty"`
}

type VersionRegisterLegacy struct {
	P              string `json:"p"`
	Op             string `json:"op"`
	SkillUID       int64  `json:"skill_uid"`
	Version        string `json:"version"`
	ManifestHRL    string `json:"manifest_hcs1"`
	Checksum       string `json:"checksum,omitempty"`
	Status         string `json:"status,omitempty"`
	Memo           string `json:"m,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type ManifestFile struct {
	Path   string `json:"path"`
	HRL    string `json:"hrl"`
	SHA256 string `json:"sha256"`
	Mime   string `json:"mime"`
}

type SkillManifest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	License     string           `json:"license"`
	Author      any              `json:"author"`
	Tags        []int64          `json:"tags,omitempty"`
	Homepage    string           `json:"homepage,omitempty"`
	Languages   []string         `json:"languages,omitempty"`
	Repo        string           `json:"repo,omitempty"`
	Commit      string           `json:"commit,omitempty"`
	Entrypoints []map[string]any `json:"entrypoints,omitempty"`
	Files       []ManifestFile   `json:"files"`
}

type ClientConfig struct {
	Network       string
	MirrorBaseURL string
	MirrorAPIKey  string
}

type TopicMemo struct {
	Protocol   string
	Indexed    bool
	TTLSeconds int64
	TopicType  TopicType
}

type TransactionMemo struct {
	Protocol  string
	Operation Operation
	TopicType TopicType
}

type ResolvedSkill struct {
	DirectoryTopicID       string
	SkillUID               int64
	Discovery              DiscoveryRegister
	VersionRegistryTopicID string
	LatestVersion          any
	Manifest               SkillManifest
	ManifestSHA256Hex      string
}

var (
	topicIDPattern  = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	hrlPattern      = regexp.MustCompile(`^hcs:\/\/1\/\d+\.\d+\.\d+$`)
	semverPattern   = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)
	checksumPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	discoveryMemoRe = regexp.MustCompile(`^hcs-26:(\d+):(\d+):(\d+)$`)
	txMemoRe        = regexp.MustCompile(`^hcs-26:op:(\d+):(\d+)$`)
)
