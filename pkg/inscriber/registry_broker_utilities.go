package inscriber

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	defaultRegistryBrokerURL      = "https://hol.org/registry/api/v1"
	defaultRegistryPollIntervalMs = int64(2000)
	defaultRegistryTimeoutMs      = int64(120000)
)

type SkillInscriptionOptions struct {
	InscribeViaRegistryBrokerOptions
	SkillName    string
	SkillVersion string
}

func InscribeViaRegistryBroker(
	ctx context.Context,
	input InscriptionInput,
	options InscribeViaRegistryBrokerOptions,
) (InscribeViaBrokerResult, error) {
	request, err := buildBrokerQuoteRequest(input, options)
	if err != nil {
		return InscribeViaBrokerResult{}, err
	}

	baseURL := strings.TrimSpace(options.BaseURL)
	if baseURL == "" {
		baseURL = defaultRegistryBrokerURL
	}

	apiKey := strings.TrimSpace(options.LedgerAPIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(options.APIKey)
	}
	if apiKey == "" {
		return InscribeViaBrokerResult{}, fmt.Errorf("either ledgerApiKey or apiKey is required for Registry Broker inscription")
	}

	brokerClient, err := NewBrokerClient(baseURL, apiKey)
	if err != nil {
		return InscribeViaBrokerResult{}, err
	}

	waitForConfirmation := boolOptionOrDefault(options.WaitForConfirmation, true)
	if !waitForConfirmation {
		job, createErr := brokerClient.CreateJob(ctx, request)
		if createErr != nil {
			return InscribeViaBrokerResult{}, createErr
		}
		jobID := strings.TrimSpace(job.JobID)
		if jobID == "" {
			jobID = strings.TrimSpace(job.ID)
		}
		return InscribeViaBrokerResult{
			Confirmed: false,
			JobID:     jobID,
			Status:    job.Status,
			HRL:       job.HRL,
			TopicID:   job.TopicID,
			Network:   job.Network,
			Error:     job.Error,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		}, nil
	}

	timeout := options.WaitTimeoutMs
	if timeout <= 0 {
		timeout = defaultRegistryTimeoutMs
	}

	if options.PollIntervalMs > 0 {
		brokerClient.pollInterval = time.Duration(options.PollIntervalMs) * time.Millisecond
	} else {
		brokerClient.pollInterval = time.Duration(defaultRegistryPollIntervalMs) * time.Millisecond
	}

	return brokerClient.InscribeAndWait(ctx, request, time.Duration(timeout)*time.Millisecond)
}

func GetRegistryBrokerQuote(
	ctx context.Context,
	input InscriptionInput,
	options InscribeViaRegistryBrokerOptions,
) (BrokerQuoteResponse, error) {
	request, err := buildBrokerQuoteRequest(input, options)
	if err != nil {
		return BrokerQuoteResponse{}, err
	}

	baseURL := strings.TrimSpace(options.BaseURL)
	if baseURL == "" {
		baseURL = defaultRegistryBrokerURL
	}

	apiKey := strings.TrimSpace(options.LedgerAPIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(options.APIKey)
	}
	if apiKey == "" {
		return BrokerQuoteResponse{}, fmt.Errorf("either ledgerApiKey or apiKey is required for Registry Broker quote")
	}

	brokerClient, err := NewBrokerClient(baseURL, apiKey)
	if err != nil {
		return BrokerQuoteResponse{}, err
	}

	return brokerClient.CreateQuote(ctx, request)
}

func InscribeSkillViaRegistryBroker(
	ctx context.Context,
	input InscriptionInput,
	options SkillInscriptionOptions,
) (InscribeViaBrokerResult, error) {
	base := options.InscribeViaRegistryBrokerOptions
	if base.Mode == "" {
		base.Mode = ModeBulkFiles
	}
	if base.Metadata == nil {
		base.Metadata = map[string]any{}
	}
	if strings.TrimSpace(options.SkillName) != "" {
		base.Metadata["skillName"] = strings.TrimSpace(options.SkillName)
	}
	if strings.TrimSpace(options.SkillVersion) != "" {
		base.Metadata["skillVersion"] = strings.TrimSpace(options.SkillVersion)
	}
	base.Metadata["kind"] = "skill"
	return InscribeViaRegistryBroker(ctx, input, base)
}

func buildBrokerQuoteRequest(
	input InscriptionInput,
	options InscribeViaRegistryBrokerOptions,
) (BrokerQuoteRequest, error) {
	mode := options.Mode
	if mode == "" {
		mode = ModeFile
	}

	request := BrokerQuoteRequest{
		Mode:         mode,
		Metadata:     options.Metadata,
		Tags:         options.Tags,
		FileStandard: strings.TrimSpace(options.FileStandard),
		ChunkSize:    options.ChunkSize,
	}

	switch input.Type {
	case InscriptionInputTypeURL:
		if strings.TrimSpace(input.URL) == "" {
			return BrokerQuoteRequest{}, fmt.Errorf("input.url is required for url input type")
		}
		request.InputType = "url"
		request.URL = strings.TrimSpace(input.URL)
	case InscriptionInputTypeFile:
		base64Value, fileName, mimeType, err := convertFilePathToBase64(input.Path)
		if err != nil {
			return BrokerQuoteRequest{}, err
		}
		request.InputType = "base64"
		request.Base64 = base64Value
		request.FileName = fileName
		request.MimeType = mimeType
	case InscriptionInputTypeBuffer:
		if len(input.Buffer) == 0 {
			return BrokerQuoteRequest{}, fmt.Errorf("input.buffer is required for buffer input type")
		}
		if strings.TrimSpace(input.FileName) == "" {
			return BrokerQuoteRequest{}, fmt.Errorf("input.fileName is required for buffer input type")
		}
		request.InputType = "base64"
		request.Base64 = encodeBufferToBase64(input.Buffer)
		request.FileName = strings.TrimSpace(input.FileName)
		request.MimeType = strings.TrimSpace(input.MimeType)
		if request.MimeType == "" {
			request.MimeType = guessMimeTypeFromName(request.FileName)
		}
	default:
		return BrokerQuoteRequest{}, fmt.Errorf("input.type must be one of: url, file, buffer")
	}

	return request, nil
}
