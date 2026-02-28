package inscriber

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Inscribe inscribes the requested payload.
func Inscribe(
	ctx context.Context,
	input InscriptionInput,
	clientConfig HederaClientConfig,
	options InscriptionOptions,
	existingClient *Client,
) (InscriptionResponse, error) {
	normalizedOptions := normalizeInscriptionOptions(options, clientConfig)
	if normalizedOptions.QuoteOnly {
		return GenerateQuote(ctx, input, clientConfig, normalizedOptions, existingClient)
	}

	client, err := resolveInscriberClient(ctx, clientConfig, normalizedOptions, existingClient)
	if err != nil {
		return InscriptionResponse{}, err
	}

	request, err := buildStartInscriptionRequest(
		input,
		clientConfig.AccountID,
		normalizedOptions.Network,
		normalizedOptions,
	)
	if err != nil {
		return InscriptionResponse{}, err
	}

	var job InscriptionJob
	switch normalizedOptions.ConnectionMode {
	case ConnectionModeWebSocket:
		job, err = client.StartInscription(ctx, request)
		if err != nil {
			return InscriptionResponse{}, err
		}
	case ConnectionModeAuto:
		job, err = client.StartInscription(ctx, request)
		if err != nil {
			return InscriptionResponse{}, err
		}
	default:
		job, err = client.StartInscription(ctx, request)
		if err != nil {
			return InscriptionResponse{}, err
		}
	}

	if strings.TrimSpace(job.TransactionBytes) == "" {
		return InscriptionResponse{}, fmt.Errorf("inscription start did not return transaction bytes")
	}

	executedTransactionID, err := ExecuteTransaction(ctx, job.TransactionBytes, clientConfig)
	if err != nil {
		return InscriptionResponse{}, err
	}

	result := InscriptionResult{
		JobID:         normalizeTransactionID(job.TxID),
		TransactionID: normalizeTransactionID(executedTransactionID),
		TopicID:       job.TopicID,
		Status:        job.Status,
		Completed:     false,
	}

	shouldWait := boolOptionOrDefault(normalizedOptions.WaitForConfirmation, true)
	if !shouldWait {
		costSummary, _ := resolveInscriptionCostSummary(
			ctx,
			executedTransactionID,
			normalizedOptions.Network,
		)
		return InscriptionResponse{
			Confirmed:   false,
			Result:      result,
			CostSummary: costSummary,
		}, nil
	}

	waited, err := waitForInscriptionWithConnection(
		ctx,
		client,
		executedTransactionID,
		normalizedOptions,
	)
	if err != nil {
		return InscriptionResponse{}, err
	}

	result.TopicID = waited.TopicID
	result.Status = waited.Status
	result.Completed = waited.Completed

	costSummary, _ := resolveInscriptionCostSummary(
		ctx,
		executedTransactionID,
		normalizedOptions.Network,
	)

	return InscriptionResponse{
		Confirmed:   waited.Completed || strings.EqualFold(waited.Status, "completed"),
		Result:      result,
		Inscription: &waited,
		CostSummary: costSummary,
	}, nil
}

// GenerateQuote performs the requested operation.
func GenerateQuote(
	ctx context.Context,
	input InscriptionInput,
	clientConfig HederaClientConfig,
	options InscriptionOptions,
	existingClient *Client,
) (InscriptionResponse, error) {
	normalizedOptions := normalizeInscriptionOptions(options, clientConfig)
	client, err := resolveInscriberClient(ctx, clientConfig, normalizedOptions, existingClient)
	if err != nil {
		return InscriptionResponse{}, err
	}

	request, err := buildStartInscriptionRequest(
		input,
		clientConfig.AccountID,
		normalizedOptions.Network,
		normalizedOptions,
	)
	if err != nil {
		return InscriptionResponse{}, err
	}

	job, err := client.StartInscription(ctx, request)
	if err != nil {
		return InscriptionResponse{}, err
	}

	quote, err := parseJobQuote(job)
	if err != nil {
		return InscriptionResponse{}, err
	}

	return InscriptionResponse{
		Confirmed: false,
		Quote:     true,
		Result:    quote,
	}, nil
}

// RetrieveInscription performs the requested operation.
func RetrieveInscription(
	ctx context.Context,
	transactionID string,
	options RetrieveInscriptionOptions,
) (InscriptionJob, error) {
	apiKey := strings.TrimSpace(options.APIKey)
	if apiKey == "" {
		accountID := strings.TrimSpace(options.AccountID)
		privateKey := strings.TrimSpace(options.PrivateKey)
		if accountID == "" || privateKey == "" {
			return InscriptionJob{}, fmt.Errorf("either API key or account ID/private key are required")
		}

		network := options.Network
		if network == "" {
			network = NetworkMainnet
		}
		authClient := NewAuthClient(options.BaseURL)
		authResult, err := authClient.Authenticate(ctx, accountID, privateKey, network)
		if err != nil {
			return InscriptionJob{}, err
		}
		apiKey = authResult.APIKey
	}

	network := options.Network
	if network == "" {
		network = NetworkMainnet
	}

	client, err := NewClient(Config{
		APIKey:  apiKey,
		Network: network,
		BaseURL: options.BaseURL,
	})
	if err != nil {
		return InscriptionJob{}, err
	}

	return client.RetrieveInscription(ctx, transactionID)
}

// WaitForInscriptionConfirmation performs the requested operation.
func WaitForInscriptionConfirmation(
	ctx context.Context,
	client *Client,
	transactionID string,
	maxAttempts int,
	intervalMs int64,
	progressCallback RegistrationProgressCallback,
) (InscriptionJob, error) {
	if client == nil {
		return InscriptionJob{}, fmt.Errorf("client is required")
	}
	options := InscriptionOptions{
		ConnectionMode:   client.connectionMode,
		WaitMaxAttempts:  maxAttempts,
		WaitInterval:     intervalMs,
		ProgressCallback: progressCallback,
	}
	return waitForInscriptionWithConnection(ctx, client, transactionID, options)
}

func waitForInscriptionWithConnection(
	ctx context.Context,
	client *Client,
	transactionID string,
	options InscriptionOptions,
) (InscriptionJob, error) {
	maxAttempts := options.WaitMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 450
	}

	intervalMs := options.WaitInterval
	if intervalMs <= 0 {
		intervalMs = 4000
	}

	switch options.ConnectionMode {
	case ConnectionModeWebSocket:
		job, err := client.waitForInscriptionWebSocket(ctx, transactionID, options.ProgressCallback)
		if err == nil {
			return job, nil
		}
		return client.WaitForInscription(ctx, transactionID, WaitOptions{
			MaxAttempts: maxAttempts,
			Interval:    time.Duration(intervalMs) * time.Millisecond,
		})
	case ConnectionModeAuto:
		job, err := client.waitForInscriptionWebSocket(ctx, transactionID, options.ProgressCallback)
		if err == nil {
			return job, nil
		}
		return client.WaitForInscription(ctx, transactionID, WaitOptions{
			MaxAttempts: maxAttempts,
			Interval:    time.Duration(intervalMs) * time.Millisecond,
		})
	default:
		return client.WaitForInscription(ctx, transactionID, WaitOptions{
			MaxAttempts: maxAttempts,
			Interval:    time.Duration(intervalMs) * time.Millisecond,
		})
	}
}

func normalizeInscriptionOptions(
	options InscriptionOptions,
	clientConfig HederaClientConfig,
) InscriptionOptions {
	normalizedOptions := options

	if normalizedOptions.Mode == "" {
		normalizedOptions.Mode = ModeFile
	}

	if normalizedOptions.ConnectionMode == "" {
		if normalizedOptions.WebSocket != nil {
			if *normalizedOptions.WebSocket {
				normalizedOptions.ConnectionMode = ConnectionModeWebSocket
			} else {
				normalizedOptions.ConnectionMode = ConnectionModeHTTP
			}
		} else {
			normalizedOptions.ConnectionMode = ConnectionModeWebSocket
		}
	}

	if normalizedOptions.Network == "" {
		if clientConfig.Network != "" {
			normalizedOptions.Network = clientConfig.Network
		} else {
			normalizedOptions.Network = NetworkMainnet
		}
	}

	return normalizedOptions
}

func resolveInscriberClient(
	ctx context.Context,
	clientConfig HederaClientConfig,
	options InscriptionOptions,
	existingClient *Client,
) (*Client, error) {
	if existingClient != nil {
		return existingClient, nil
	}

	apiKey := strings.TrimSpace(options.APIKey)
	if apiKey == "" {
		network := options.Network
		if network == "" {
			network = NetworkMainnet
		}
		authClient := NewAuthClient(options.BaseURL)
		authResult, err := authClient.Authenticate(
			ctx,
			clientConfig.AccountID,
			clientConfig.PrivateKey,
			network,
		)
		if err != nil {
			return nil, err
		}
		apiKey = authResult.APIKey
	}

	return NewClient(Config{
		APIKey:         apiKey,
		Network:        options.Network,
		BaseURL:        options.BaseURL,
		ConnectionMode: options.ConnectionMode,
	})
}

func boolOptionOrDefault(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}
