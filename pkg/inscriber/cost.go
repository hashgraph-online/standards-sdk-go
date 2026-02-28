package inscriber

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/mirror"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

const tinybarDivisor = 100000000.0

func parseJobQuote(job InscriptionJob) (QuoteResult, error) {
	totalCostHBAR := "0.001"

	if job.TotalCost > 0 {
		totalCostHBAR = formatTinybarToHBAR(job.TotalCost)
	} else if strings.TrimSpace(job.TransactionBytes) != "" {
		decoded, err := base64.StdEncoding.DecodeString(job.TransactionBytes)
		if err == nil {
			transferTinybar, transferErr := parseTransferTinybar(decoded)
			if transferErr == nil && transferTinybar > 0 {
				totalCostHBAR = formatTinybarToHBAR(transferTinybar)
			}
		}
	}

	quote := QuoteResult{
		TotalCostHBAR: totalCostHBAR,
		ValidUntil:    time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339),
	}
	quote.Breakdown.Transfers = []QuoteTransfer{
		{
			To:          "Inscription Service",
			Amount:      totalCostHBAR,
			Description: "Inscription fee",
		},
	}

	return quote, nil
}

func resolveInscriptionCostSummary(
	ctx context.Context,
	transactionID string,
	network Network,
) (*InscriptionCostSummary, error) {
	normalizedTxID := normalizeTransactionID(transactionID)
	if strings.TrimSpace(normalizedTxID) == "" {
		return nil, nil
	}

	networkValue := string(network)
	if strings.TrimSpace(networkValue) == "" {
		networkValue = string(NetworkMainnet)
	}

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: networkValue,
	})
	if err != nil {
		return nil, err
	}

	transaction, err := mirrorClient.GetTransaction(ctx, normalizedTxID)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, nil
	}

	payer := normalizedPayerAccount(normalizedTxID)
	totalTinybar := int64(0)
	if payer != "" {
		for _, transfer := range transaction.Transfers {
			if transfer.Account == payer && transfer.Amount < 0 {
				totalTinybar = absInt64(transfer.Amount)
				break
			}
		}
	}
	if totalTinybar <= 0 {
		for _, transfer := range transaction.Transfers {
			if transfer.Amount > 0 {
				totalTinybar += transfer.Amount
			}
		}
	}
	if totalTinybar <= 0 && transaction.ChargedTxFee > 0 {
		totalTinybar = transaction.ChargedTxFee
	}
	if totalTinybar <= 0 {
		return nil, nil
	}

	summary := &InscriptionCostSummary{
		TotalCostHBAR: formatTinybarToHBAR(totalTinybar),
	}

	breakdownTransfers := make([]QuoteTransfer, 0)
	for _, transfer := range transaction.Transfers {
		if transfer.Amount <= 0 {
			continue
		}
		breakdownTransfers = append(breakdownTransfers, QuoteTransfer{
			To:          transfer.Account,
			Amount:      formatTinybarToHBAR(transfer.Amount),
			Description: fmt.Sprintf("HBAR transfer from %s", payer),
		})
	}

	if len(breakdownTransfers) == 0 {
		breakdownTransfers = append(breakdownTransfers, QuoteTransfer{
			To:          "Hedera network",
			Amount:      summary.TotalCostHBAR,
			Description: fmt.Sprintf("Transaction fee debited from %s", payer),
		})
	}

	summary.Breakdown.Transfers = breakdownTransfers
	return summary, nil
}

func parseTransferTinybar(transactionBytes []byte) (int64, error) {
	decodedTransaction, err := transactionFromBytes(transactionBytes)
	if err != nil {
		return 0, err
	}
	transferTransaction, transferErr := asTransferTransaction(decodedTransaction)
	if transferErr != nil {
		return 0, transferErr
	}

	totalTinybar := int64(0)
	for _, amount := range transferTransaction.GetHbarTransfers() {
		if amount.AsTinybar() < 0 {
			totalTinybar += absInt64(amount.AsTinybar())
		}
	}
	if totalTinybar > 0 {
		return totalTinybar, nil
	}

	feeTinybar := transferTransaction.GetMaxTransactionFee().AsTinybar()
	if feeTinybar > 0 {
		return feeTinybar, nil
	}

	return 0, fmt.Errorf("unable to compute transfer tinybar amount")
}

func transactionFromBytes(transactionBytes []byte) (any, error) {
	decodedTransaction, err := hedera.TransactionFromBytes(transactionBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction bytes: %w", err)
	}
	return decodedTransaction, nil
}

func formatTinybarToHBAR(tinybar int64) string {
	value := float64(tinybar) / tinybarDivisor
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func normalizedPayerAccount(transactionID string) string {
	parts := strings.Split(strings.TrimSpace(transactionID), "-")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
