package hcs11

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func (c *Client) UpdateAccountMemoWithProfile(
	ctx context.Context,
	accountID string,
	profileTopicID string,
) (TransactionResult, error) {
	if strings.TrimSpace(accountID) == "" {
		return TransactionResult{
			Success: false,
			Error:   "account ID is required",
		}, nil
	}
	if strings.TrimSpace(c.operatorPrivateKey) == "" {
		return TransactionResult{
			Success: false,
			Error:   "operator private key is required",
		}, nil
	}

	parsedAccountID, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		return TransactionResult{}, fmt.Errorf("invalid account ID: %w", err)
	}

	transaction := hedera.NewAccountUpdateTransaction().
		SetAccountID(parsedAccountID).
		SetAccountMemo(c.SetProfileForAccountMemo(profileTopicID, 1))

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return TransactionResult{}, err
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return TransactionResult{}, err
	}
	if receipt.Status.String() != "SUCCESS" {
		return TransactionResult{
			Success: false,
			Error:   fmt.Sprintf("transaction failed: %s", receipt.Status.String()),
		}, nil
	}
	return TransactionResult{
		Success: true,
	}, nil
}

func (c *Client) FetchProfileByAccountID(
	ctx context.Context,
	accountID string,
	network string,
) (FetchProfileResponse, error) {
	normalizedAccountID := strings.TrimSpace(accountID)
	if normalizedAccountID == "" {
		return FetchProfileResponse{
			Success: false,
			Error:   "account ID is required",
		}, nil
	}
	if strings.TrimSpace(network) == "" {
		network = c.network
	}

	memo, err := c.mirrorClient.GetAccountMemo(ctx, normalizedAccountID)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	if !strings.HasPrefix(memo, "hcs-11:") {
		return FetchProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("account %s does not have a valid HCS-11 memo", normalizedAccountID),
		}, nil
	}

	reference := strings.TrimPrefix(memo, "hcs-11:")
	switch {
	case strings.HasPrefix(reference, "hcs://"):
		return c.fetchFromHCSReference(ctx, reference, network)
	case strings.HasPrefix(reference, "ipfs://"):
		ipfsHash := strings.TrimPrefix(reference, "ipfs://")
		return c.fetchFromURL(ctx, "https://ipfs.io/ipfs/"+ipfsHash, "")
	case strings.HasPrefix(reference, "ar://"):
		arweaveID := strings.TrimPrefix(reference, "ar://")
		return c.fetchFromURL(ctx, "https://arweave.net/"+arweaveID, "")
	default:
		return FetchProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid protocol reference format: %s", reference),
		}, nil
	}
}

func (c *Client) fetchFromHCSReference(
	ctx context.Context,
	reference string,
	network string,
) (FetchProfileResponse, error) {
	parts := strings.Split(reference, "/")
	if len(parts) < 4 {
		return FetchProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid HCS protocol reference format: %s", reference),
		}, nil
	}
	profileTopicID := strings.TrimSpace(parts[3])
	cdnURL := fmt.Sprintf(
		"%s/api/inscription-cdn/%s?network=%s",
		c.kiloScribeBaseURL,
		profileTopicID,
		strings.TrimSpace(network),
	)

	response, err := c.fetchFromURL(ctx, cdnURL, profileTopicID)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	return response, nil
}

func (c *Client) fetchFromURL(
	ctx context.Context,
	endpoint string,
	profileTopicID string,
) (FetchProfileResponse, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return FetchProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to fetch profile content: %s", response.Status),
		}, nil
	}

	var profile HCS11Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return FetchProfileResponse{
			Success: false,
			Error:   "invalid HCS-11 profile data",
		}, nil
	}

	validation := c.ValidateProfile(profile)
	if !validation.Valid {
		return FetchProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid HCS-11 profile data: %s", strings.Join(validation.Errors, ", ")),
		}, nil
	}

	return FetchProfileResponse{
		Success: true,
		Profile: &profile,
		TopicInfo: &ResolvedTopicInfo{
			InboundTopic:   profile.InboundTopicID,
			OutboundTopic:  profile.OutboundTopicID,
			ProfileTopicID: profileTopicID,
		},
	}, nil
}
