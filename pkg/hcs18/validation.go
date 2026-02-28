package hcs18

import (
	"fmt"
	"strings"
)

// ValidateMessage performs the requested operation.
func ValidateMessage(message DiscoveryMessage) error {
	if strings.TrimSpace(message.P) != "hcs-18" {
		return fmt.Errorf("message p must be hcs-18")
	}
	switch message.Op {
	case OperationAnnounce:
		data, ok := message.Data.(AnnounceData)
		if !ok {
			return fmt.Errorf("announce data is invalid")
		}
		if strings.TrimSpace(data.Account) == "" {
			return fmt.Errorf("announce account is required")
		}
		if strings.TrimSpace(data.Petal.Name) == "" {
			return fmt.Errorf("announce petal.name is required")
		}
		if len(data.Capabilities.Protocols) == 0 {
			return fmt.Errorf("announce capabilities.protocols is required")
		}
	case OperationPropose:
		data, ok := message.Data.(ProposeData)
		if !ok {
			return fmt.Errorf("propose data is invalid")
		}
		if strings.TrimSpace(data.Proposer) == "" {
			return fmt.Errorf("propose proposer is required")
		}
		if len(data.Members) == 0 {
			return fmt.Errorf("propose members are required")
		}
		if strings.TrimSpace(data.Config.Name) == "" {
			return fmt.Errorf("propose config.name is required")
		}
		if data.Config.Threshold <= 0 {
			return fmt.Errorf("propose config.threshold must be positive")
		}
	case OperationRespond:
		data, ok := message.Data.(RespondData)
		if !ok {
			return fmt.Errorf("respond data is invalid")
		}
		if strings.TrimSpace(data.Responder) == "" {
			return fmt.Errorf("respond responder is required")
		}
		if data.ProposalSeq <= 0 {
			return fmt.Errorf("respond proposal_seq must be positive")
		}
		if data.Decision != "accept" && data.Decision != "reject" {
			return fmt.Errorf("respond decision must be accept or reject")
		}
	case OperationComplete:
		data, ok := message.Data.(CompleteData)
		if !ok {
			return fmt.Errorf("complete data is invalid")
		}
		if data.ProposalSeq <= 0 {
			return fmt.Errorf("complete proposal_seq must be positive")
		}
		if strings.TrimSpace(data.FloraAccount) == "" {
			return fmt.Errorf("complete flora_account is required")
		}
		if strings.TrimSpace(data.Topics.Communication) == "" || strings.TrimSpace(data.Topics.Transaction) == "" || strings.TrimSpace(data.Topics.State) == "" {
			return fmt.Errorf("complete topics communication/transaction/state are required")
		}
	case OperationWithdraw:
		data, ok := message.Data.(WithdrawData)
		if !ok {
			return fmt.Errorf("withdraw data is invalid")
		}
		if strings.TrimSpace(data.Account) == "" {
			return fmt.Errorf("withdraw account is required")
		}
		if data.AnnounceSeq <= 0 {
			return fmt.Errorf("withdraw announce_seq must be positive")
		}
	default:
		return fmt.Errorf("unsupported operation %q", message.Op)
	}
	return nil
}
