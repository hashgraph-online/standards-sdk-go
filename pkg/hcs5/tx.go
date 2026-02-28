package hcs5

import (
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func BuildMintTx(tokenID string, metadata string, transactionMemo string) (*hedera.TokenMintTransaction, error) {
	trimmedTokenID := strings.TrimSpace(tokenID)
	if trimmedTokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}
	parsedTokenID, err := hedera.TokenIDFromString(trimmedTokenID)
	if err != nil {
		return nil, fmt.Errorf("invalid token ID: %w", err)
	}

	transaction := hedera.NewTokenMintTransaction().
		SetTokenID(parsedTokenID).
		SetMetadata([]byte(metadata))

	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(transactionMemo)
	}

	return transaction, nil
}

func BuildMintWithHRLTx(
	tokenID string,
	metadataTopicID string,
	transactionMemo string,
) (*hedera.TokenMintTransaction, error) {
	trimmedTopicID := strings.TrimSpace(metadataTopicID)
	if trimmedTopicID == "" {
		return nil, fmt.Errorf("metadata topic ID is required")
	}

	return BuildMintTx(tokenID, BuildHCS1HRL(trimmedTopicID), transactionMemo)
}
