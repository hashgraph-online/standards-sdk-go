package hcs18

type DiscoveryOperation string

const (
	OperationAnnounce DiscoveryOperation = "announce"
	OperationPropose  DiscoveryOperation = "propose"
	OperationRespond  DiscoveryOperation = "respond"
	OperationComplete DiscoveryOperation = "complete"
	OperationWithdraw DiscoveryOperation = "withdraw"
)

type AnnounceData struct {
	Account      string            `json:"account"`
	Petal        PetalDescriptor   `json:"petal"`
	Capabilities CapabilityDetails `json:"capabilities"`
	ValidFor     int64             `json:"valid_for,omitempty"`
}

type PetalDescriptor struct {
	Name     string `json:"name"`
	Priority int64  `json:"priority"`
}

type CapabilityDetails struct {
	Protocols        []string          `json:"protocols"`
	Resources        map[string]string `json:"resources,omitempty"`
	GroupPreferences map[string]any    `json:"group_preferences,omitempty"`
}

type ProposeMember struct {
	Account     string `json:"account"`
	AnnounceSeq int64  `json:"announce_seq,omitempty"`
	Priority    int64  `json:"priority"`
	Status      string `json:"status,omitempty"`
}

type ProposeConfig struct {
	Name      string `json:"name"`
	Threshold int64  `json:"threshold"`
	Purpose   string `json:"purpose,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type ProposeData struct {
	Proposer      string          `json:"proposer"`
	Members       []ProposeMember `json:"members"`
	Config        ProposeConfig   `json:"config"`
	ExistingFlora string          `json:"existing_flora,omitempty"`
}

type RespondData struct {
	Responder   string `json:"responder"`
	ProposalSeq int64  `json:"proposal_seq"`
	Decision    string `json:"decision"`
	Reason      string `json:"reason,omitempty"`
	AcceptedSeq int64  `json:"accepted_seq,omitempty"`
}

type CompleteData struct {
	ProposalSeq  int64         `json:"proposal_seq"`
	FloraAccount string        `json:"flora_account"`
	Topics       CompleteTopic `json:"topics"`
	Proposer     string        `json:"proposer,omitempty"`
}

type CompleteTopic struct {
	Communication string `json:"communication"`
	Transaction   string `json:"transaction"`
	State         string `json:"state"`
}

type WithdrawData struct {
	Account     string `json:"account"`
	AnnounceSeq int64  `json:"announce_seq"`
	Reason      string `json:"reason,omitempty"`
}

type DiscoveryMessage struct {
	P    string             `json:"p"`
	Op   DiscoveryOperation `json:"op"`
	Data any                `json:"data"`
}

type MessageRecord struct {
	Message            DiscoveryMessage `json:"message"`
	ConsensusTimestamp string           `json:"consensus_timestamp"`
	SequenceNumber     int64            `json:"sequence_number"`
	Payer              string           `json:"payer"`
}

type TrackedResponse struct {
	Decision string
}

type TrackedProposal struct {
	Data      ProposeData
	Responses map[string]TrackedResponse
}

type TopicMemo struct {
	Protocol string
	Type     int
	TTL      int64
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateDiscoveryTopicOptions struct {
	TTLSeconds          int64
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
	MemoOverride        string
}
