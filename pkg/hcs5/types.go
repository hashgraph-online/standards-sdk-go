package hcs5

import "github.com/hashgraph-online/go-sdk/pkg/inscriber"

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	InscriberAuthURL   string
	InscriberAPIURL    string
}

type MintOptions struct {
	TokenID         string
	MetadataTopicID string
	SupplyKey       string
	Memo            string
}

type CreateHashinalOptions struct {
	TokenID           string
	Request           inscriber.StartInscriptionRequest
	WaitForCompletion bool
	SupplyKey         string
	Memo              string
	InscriberNetwork  inscriber.Network
	InscriberAuthURL  string
	InscriberAPIURL   string
}

type MintResponse struct {
	Success       bool
	SerialNumber  int64
	TransactionID string
	Metadata      string
	Error         string
}

func BuildHCS1HRL(topicID string) string {
	return "hcs://1/" + topicID
}
