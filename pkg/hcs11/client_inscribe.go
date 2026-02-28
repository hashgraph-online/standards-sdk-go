package hcs11

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/inscriber"
)

func (c *Client) InscribeImage(
	ctx context.Context,
	buffer []byte,
	fileName string,
	options InscribeImageOptions,
) (InscribeImageResponse, error) {
	if len(buffer) == 0 {
		return InscribeImageResponse{
			Success: false,
			Error:   "image buffer is required",
		}, nil
	}
	if strings.TrimSpace(fileName) == "" {
		return InscribeImageResponse{
			Success: false,
			Error:   "file name is required",
		}, nil
	}

	waitForConfirmation := true
	if !options.WaitForConfirmation {
		waitForConfirmation = true
	}

	authResult, network, err := c.authenticateInscriber(ctx)
	if err != nil {
		return InscribeImageResponse{}, err
	}

	inscriberClient, err := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: c.inscriberAPIURL,
	})
	if err != nil {
		return InscribeImageResponse{}, err
	}

	started, err := inscriberClient.StartInscription(ctx, inscriber.StartInscriptionRequest{
		File: inscriber.FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString(buffer),
			FileName: fileName,
			MimeType: "application/octet-stream",
		},
		HolderID: c.operatorAccountID,
		Mode:     inscriber.ModeFile,
	})
	if err != nil {
		return InscribeImageResponse{}, err
	}

	executedTransactionID, err := inscriber.ExecuteTransaction(
		ctx,
		started.TransactionBytes,
		inscriber.HederaClientConfig{
			AccountID:  c.operatorAccountID,
			PrivateKey: c.operatorPrivateKey,
			Network:    network,
		},
	)
	if err != nil {
		return InscribeImageResponse{}, err
	}

	if !waitForConfirmation {
		return InscribeImageResponse{
			Success:       false,
			TransactionID: executedTransactionID,
			Error:         "inscription not confirmed",
		}, nil
	}

	waited, err := inscriberClient.WaitForInscription(ctx, executedTransactionID, inscriber.WaitOptions{
		MaxAttempts: 150,
		Interval:    2 * time.Second,
	})
	if err != nil {
		return InscribeImageResponse{}, err
	}
	if !waited.Completed && !strings.EqualFold(waited.Status, "completed") {
		return InscribeImageResponse{
			Success:       false,
			TransactionID: executedTransactionID,
			Error:         "inscription not confirmed",
		}, nil
	}

	return InscribeImageResponse{
		ImageTopicID:  waited.TopicID,
		TransactionID: executedTransactionID,
		Success:       true,
	}, nil
}

func (c *Client) InscribeProfile(
	ctx context.Context,
	profile HCS11Profile,
	options InscribeProfileOptions,
) (InscribeProfileResponse, error) {
	waitForConfirmation := true
	if !options.WaitForConfirmation {
		waitForConfirmation = true
	}

	if err := c.AttachUAIDIfMissing(ctx, &profile); err != nil {
		return InscribeProfileResponse{}, err
	}

	validation := c.ValidateProfile(profile)
	if !validation.Valid {
		return InscribeProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid profile: %s", strings.Join(validation.Errors, ", ")),
		}, nil
	}

	profileJSON, err := c.ProfileToJSONString(profile)
	if err != nil {
		return InscribeProfileResponse{}, err
	}

	authResult, network, err := c.authenticateInscriber(ctx)
	if err != nil {
		return InscribeProfileResponse{}, err
	}

	inscriberClient, err := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: c.inscriberAPIURL,
	})
	if err != nil {
		return InscribeProfileResponse{}, err
	}

	fileName := fmt.Sprintf(
		"profile-%s.json",
		strings.ReplaceAll(strings.ToLower(strings.TrimSpace(profile.DisplayName)), " ", "-"),
	)
	started, err := inscriberClient.StartInscription(ctx, inscriber.StartInscriptionRequest{
		File: inscriber.FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString([]byte(profileJSON)),
			FileName: fileName,
			MimeType: "application/json",
		},
		HolderID: c.operatorAccountID,
		Mode:     inscriber.ModeFile,
	})
	if err != nil {
		return InscribeProfileResponse{}, err
	}

	executedTransactionID, err := inscriber.ExecuteTransaction(
		ctx,
		started.TransactionBytes,
		inscriber.HederaClientConfig{
			AccountID:  c.operatorAccountID,
			PrivateKey: c.operatorPrivateKey,
			Network:    network,
		},
	)
	if err != nil {
		return InscribeProfileResponse{}, err
	}

	if !waitForConfirmation {
		return InscribeProfileResponse{
			ProfileTopicID: "",
			TransactionID:  executedTransactionID,
			Success:        false,
			Error:          "profile inscription not confirmed",
		}, nil
	}

	waited, err := inscriberClient.WaitForInscription(ctx, executedTransactionID, inscriber.WaitOptions{
		MaxAttempts: 100,
		Interval:    2 * time.Second,
	})
	if err != nil {
		return InscribeProfileResponse{}, err
	}
	if !waited.Completed && !strings.EqualFold(waited.Status, "completed") {
		return InscribeProfileResponse{
			Success:       false,
			TransactionID: executedTransactionID,
			Error:         "failed to inscribe profile content",
		}, nil
	}

	return InscribeProfileResponse{
		ProfileTopicID:  waited.TopicID,
		TransactionID:   executedTransactionID,
		Success:         true,
		InboundTopicID:  profile.InboundTopicID,
		OutboundTopicID: profile.OutboundTopicID,
	}, nil
}

func (c *Client) CreateAndInscribeProfile(
	ctx context.Context,
	profile HCS11Profile,
	updateAccountMemo bool,
	options InscribeProfileOptions,
) (InscribeProfileResponse, error) {
	inscriptionResult, err := c.InscribeProfile(ctx, profile, options)
	if err != nil {
		return InscribeProfileResponse{}, err
	}
	if !inscriptionResult.Success {
		return inscriptionResult, nil
	}

	if updateAccountMemo {
		updateResult, updateErr := c.UpdateAccountMemoWithProfile(ctx, c.operatorAccountID, inscriptionResult.ProfileTopicID)
		if updateErr != nil {
			return InscribeProfileResponse{}, updateErr
		}
		if !updateResult.Success {
			return InscribeProfileResponse{
				ProfileTopicID: inscriptionResult.ProfileTopicID,
				TransactionID:  inscriptionResult.TransactionID,
				Success:        false,
				Error:          updateResult.Error,
			}, nil
		}
	}

	return inscriptionResult, nil
}

func (c *Client) authenticateInscriber(
	ctx context.Context,
) (inscriber.AuthResult, inscriber.Network, error) {
	if strings.TrimSpace(c.operatorAccountID) == "" || strings.TrimSpace(c.operatorPrivateKey) == "" {
		return inscriber.AuthResult{}, "", fmt.Errorf("operator credentials are required for inscription")
	}

	network := inscriber.NetworkTestnet
	if strings.EqualFold(c.network, "mainnet") {
		network = inscriber.NetworkMainnet
	}

	authClient := inscriber.NewAuthClient(c.inscriberAuthURL)
	authResult, err := authClient.Authenticate(
		ctx,
		c.operatorAccountID,
		c.operatorPrivateKey,
		network,
	)
	if err != nil {
		return inscriber.AuthResult{}, "", err
	}
	return authResult, network, nil
}
