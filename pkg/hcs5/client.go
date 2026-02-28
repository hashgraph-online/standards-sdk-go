package hcs5

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient       *hedera.Client
	operatorAccountID  hedera.AccountID
	operatorPrivateKey hedera.PrivateKey
	network            string
	inscriberAuthURL   string
	inscriberAPIURL    string
}

func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.OperatorAccountID) == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	if strings.TrimSpace(config.OperatorPrivateKey) == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	accountID, err := hedera.AccountIDFromString(config.OperatorAccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	privateKey, err := shared.ParsePrivateKey(config.OperatorPrivateKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(accountID, privateKey)

	return &Client{
		hederaClient:       hederaClient,
		operatorAccountID:  accountID,
		operatorPrivateKey: privateKey,
		network:            network,
		inscriberAuthURL:   strings.TrimSpace(config.InscriberAuthURL),
		inscriberAPIURL:    strings.TrimSpace(config.InscriberAPIURL),
	}, nil
}

func (c *Client) Mint(
	ctx context.Context,
	options MintOptions,
) (MintResponse, error) {
	if strings.TrimSpace(options.MetadataTopicID) == "" {
		return MintResponse{
			Success: false,
			Error:   "metadataTopicID is required",
		}, nil
	}

	transaction, err := BuildMintWithHRLTx(options.TokenID, options.MetadataTopicID, options.Memo)
	if err != nil {
		return MintResponse{}, err
	}

	frozenTransaction, err := transaction.FreezeWith(c.hederaClient)
	if err != nil {
		return MintResponse{}, fmt.Errorf("failed to freeze mint transaction: %w", err)
	}

	if strings.TrimSpace(options.SupplyKey) != "" {
		supplyKey, parseErr := shared.ParsePrivateKey(options.SupplyKey)
		if parseErr != nil {
			return MintResponse{}, parseErr
		}
		frozenTransaction = frozenTransaction.Sign(supplyKey)
	}

	response, err := hedera.TransactionExecute(frozenTransaction, c.hederaClient)
	if err != nil {
		return MintResponse{}, fmt.Errorf("failed to execute mint transaction: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return MintResponse{}, fmt.Errorf("failed to retrieve mint receipt: %w", err)
	}
	if receipt.Status.String() != "SUCCESS" {
		return MintResponse{}, fmt.Errorf("mint transaction failed with status %s", receipt.Status.String())
	}

	serial := int64(0)
	if len(receipt.SerialNumbers) > 0 {
		serial = receipt.SerialNumbers[0]
	}

	return MintResponse{
		Success:       true,
		SerialNumber:  serial,
		TransactionID: response.TransactionID.String(),
		Metadata:      BuildHCS1HRL(options.MetadataTopicID),
	}, nil
}

func (c *Client) CreateHashinal(
	ctx context.Context,
	options CreateHashinalOptions,
) (MintResponse, error) {
	request := options.Request
	if request.Mode == "" {
		request.Mode = inscriber.ModeHashinal
	}
	if request.HolderID == "" {
		request.HolderID = c.operatorAccountID.String()
	}

	inscriberNetwork := options.InscriberNetwork
	if inscriberNetwork == "" {
		if c.network == shared.NetworkMainnet {
			inscriberNetwork = inscriber.NetworkMainnet
		} else {
			inscriberNetwork = inscriber.NetworkTestnet
		}
	}

	authBaseURL := options.InscriberAuthURL
	if strings.TrimSpace(authBaseURL) == "" {
		authBaseURL = c.inscriberAuthURL
	}
	apiBaseURL := options.InscriberAPIURL
	if strings.TrimSpace(apiBaseURL) == "" {
		apiBaseURL = c.inscriberAPIURL
	}

	authClient := inscriber.NewAuthClient(authBaseURL)
	authResult, err := authClient.Authenticate(
		ctx,
		c.operatorAccountID.String(),
		c.operatorPrivateKey.String(),
		inscriberNetwork,
	)
	if err != nil {
		return MintResponse{}, fmt.Errorf("failed to authenticate inscriber client: %w", err)
	}

	inscriberClient, err := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: inscriberNetwork,
		BaseURL: apiBaseURL,
	})
	if err != nil {
		return MintResponse{}, err
	}

	job, err := inscriberClient.StartInscription(ctx, request)
	if err != nil {
		return MintResponse{}, fmt.Errorf("failed to start inscription: %w", err)
	}
	if strings.TrimSpace(job.TransactionBytes) == "" {
		return MintResponse{}, fmt.Errorf("inscriber response did not include transaction bytes")
	}

	executedTransactionID, err := inscriber.ExecuteTransaction(
		ctx,
		job.TransactionBytes,
		inscriber.HederaClientConfig{
			AccountID:  c.operatorAccountID.String(),
			PrivateKey: c.operatorPrivateKey.String(),
			Network:    inscriberNetwork,
		},
	)
	if err != nil {
		return MintResponse{}, err
	}

	waitForCompletion := options.WaitForCompletion
	if !waitForCompletion {
		waitForCompletion = true
	}

	if waitForCompletion {
		waited, waitErr := inscriberClient.WaitForInscription(ctx, executedTransactionID, inscriber.WaitOptions{
			MaxAttempts: 90,
			Interval:    2 * time.Second,
		})
		if waitErr != nil {
			return MintResponse{}, waitErr
		}
		if !waited.Completed && !strings.EqualFold(waited.Status, "completed") {
			return MintResponse{}, fmt.Errorf("inscription did not complete successfully")
		}
		if strings.TrimSpace(waited.TopicID) == "" {
			return MintResponse{}, fmt.Errorf("inscription completion did not include topic ID")
		}
		return c.Mint(ctx, MintOptions{
			TokenID:         options.TokenID,
			MetadataTopicID: waited.TopicID,
			SupplyKey:       options.SupplyKey,
			Memo:            options.Memo,
		})
	}

	if strings.TrimSpace(job.TopicID) == "" {
		decodedBytes, decodeErr := base64.StdEncoding.DecodeString(job.TransactionBytes)
		if decodeErr == nil && len(decodedBytes) > 0 {
			return MintResponse{}, fmt.Errorf("inscription not completed; unable to mint without topic ID")
		}
		return MintResponse{}, fmt.Errorf("inscription not completed; topic ID unavailable")
	}

	return c.Mint(ctx, MintOptions{
		TokenID:         options.TokenID,
		MetadataTopicID: job.TopicID,
		SupplyKey:       options.SupplyKey,
		Memo:            options.Memo,
	})
}
