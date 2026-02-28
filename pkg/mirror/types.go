package mirror

type TopicInfo struct {
	AdminKey         map[string]any   `json:"admin_key"`
	AutoRenewAccount string           `json:"auto_renew_account"`
	AutoRenewPeriod  int64            `json:"auto_renew_period"`
	CreatedTimestamp string           `json:"created_timestamp"`
	Deleted          bool             `json:"deleted"`
	Memo             string           `json:"memo"`
	SubmitKey        map[string]any   `json:"submit_key"`
	TopicID          string           `json:"topic_id"`
	FeeScheduleKey   map[string]any   `json:"fee_schedule_key"`
	FeeExemptKeyList []map[string]any `json:"fee_exempt_key_list"`
}

type AccountInfo struct {
	Account string         `json:"account"`
	Key     map[string]any `json:"key"`
	Memo    string         `json:"memo"`
}

type TopicMessage struct {
	ConsensusTimestamp string     `json:"consensus_timestamp"`
	ChunkInfo          *ChunkInfo `json:"chunk_info,omitempty"`
	Message            string     `json:"message"`
	PayerAccountID     string     `json:"payer_account_id"`
	RunningHash        string     `json:"running_hash"`
	RunningHashVersion int64      `json:"running_hash_version"`
	SequenceNumber     int64      `json:"sequence_number"`
	TopicID            string     `json:"topic_id"`
}

type ChunkInfo struct {
	InitialTransactionID any `json:"initial_transaction_id,omitempty"`
	Number               int `json:"number,omitempty"`
	Total                int `json:"total,omitempty"`
}

type topicMessagesResponse struct {
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
	Messages []TopicMessage `json:"messages"`
}

type Transaction struct {
	ChargedTxFee       int64      `json:"charged_tx_fee"`
	ConsensusTimestamp string     `json:"consensus_timestamp"`
	EntityID           *string    `json:"entity_id"`
	MaxFee             string     `json:"max_fee"`
	MemoBase64         string     `json:"memo_base64"`
	Name               string     `json:"name"`
	Node               string     `json:"node"`
	Result             string     `json:"result"`
	TransactionID      string     `json:"transaction_id"`
	Transfers          []Transfer `json:"transfers"`
}

type Transfer struct {
	Account    string `json:"account"`
	Amount     int64  `json:"amount"`
	IsApproval bool   `json:"is_approval"`
}

type transactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	Links        struct {
		Next string `json:"next"`
	} `json:"links"`
}
