package inscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BrokerClient struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	pollInterval time.Duration
}

// NewBrokerClient creates a new BrokerClient.
func NewBrokerClient(baseURL string, apiKey string) (*BrokerClient, error) {
	normalizedURL := strings.TrimSpace(baseURL)
	if normalizedURL == "" {
		normalizedURL = "https://registry.hashgraphonline.com/api/v1"
	}
	normalizedURL = strings.TrimRight(normalizedURL, "/")

	normalizedAPIKey := strings.TrimSpace(apiKey)
	if normalizedAPIKey == "" {
		return nil, fmt.Errorf("registry broker API key is required")
	}

	return &BrokerClient{
		baseURL:      normalizedURL,
		apiKey:       normalizedAPIKey,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		pollInterval: 2 * time.Second,
	}, nil
}

// CreateQuote creates the requested resource.
func (c *BrokerClient) CreateQuote(
	ctx context.Context,
	request BrokerQuoteRequest,
) (BrokerQuoteResponse, error) {
	var response BrokerQuoteResponse
	if err := c.postJSON(ctx, "/inscribe/content/quote", request, &response); err != nil {
		return BrokerQuoteResponse{}, err
	}
	return response, nil
}

// CreateJob creates the requested resource.
func (c *BrokerClient) CreateJob(
	ctx context.Context,
	request BrokerQuoteRequest,
) (BrokerJobResponse, error) {
	var response BrokerJobResponse
	if err := c.postJSON(ctx, "/inscribe/content", request, &response); err != nil {
		return BrokerJobResponse{}, err
	}
	return response, nil
}

// GetJob returns the requested value.
func (c *BrokerClient) GetJob(
	ctx context.Context,
	jobID string,
) (BrokerJobResponse, error) {
	var response BrokerJobResponse
	if err := c.getJSON(ctx, "/inscribe/content/"+jobID, &response); err != nil {
		return BrokerJobResponse{}, err
	}
	return response, nil
}

// WaitForJob performs the requested operation.
func (c *BrokerClient) WaitForJob(
	ctx context.Context,
	jobID string,
	timeout time.Duration,
) (BrokerJobResponse, error) {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	deadline := time.Now().Add(timeout)
	var latest BrokerJobResponse

	for time.Now().Before(deadline) {
		job, err := c.GetJob(ctx, jobID)
		if err != nil {
			return BrokerJobResponse{}, err
		}
		latest = job

		if strings.EqualFold(job.Status, "completed") {
			return latest, nil
		}
		if strings.EqualFold(job.Status, "failed") {
			if job.Error == "" {
				job.Error = "registry broker inscription failed"
			}
			return latest, errors.New(job.Error)
		}

		select {
		case <-ctx.Done():
			return BrokerJobResponse{}, ctx.Err()
		case <-time.After(c.pollInterval):
		}
	}

	return latest, fmt.Errorf("registry broker job %s did not complete before timeout", jobID)
}

// InscribeAndWait inscribes the requested payload.
func (c *BrokerClient) InscribeAndWait(
	ctx context.Context,
	request BrokerQuoteRequest,
	timeout time.Duration,
) (InscribeViaBrokerResult, error) {
	job, err := c.CreateJob(ctx, request)
	if err != nil {
		return InscribeViaBrokerResult{}, err
	}

	jobID := job.JobID
	if jobID == "" {
		jobID = job.ID
	}
	if jobID == "" {
		return InscribeViaBrokerResult{}, fmt.Errorf("registry broker response missing job ID")
	}

	finalJob, err := c.WaitForJob(ctx, jobID, timeout)
	if err != nil {
		return InscribeViaBrokerResult{}, err
	}

	return InscribeViaBrokerResult{
		Confirmed: strings.EqualFold(finalJob.Status, "completed"),
		JobID:     jobID,
		Status:    finalJob.Status,
		HRL:       finalJob.HRL,
		TopicID:   finalJob.TopicID,
		Network:   finalJob.Network,
		Error:     finalJob.Error,
		CreatedAt: finalJob.CreatedAt,
		UpdatedAt: finalJob.UpdatedAt,
	}, nil
}

func (c *BrokerClient) postJSON(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-api-key", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"registry broker POST %s failed with status %d: %s",
			path,
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	if err := json.Unmarshal(responseBody, target); err != nil {
		return fmt.Errorf("failed to decode registry broker response: %w", err)
	}

	return nil
}

func (c *BrokerClient) getJSON(ctx context.Context, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-api-key", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"registry broker GET %s failed with status %d: %s",
			path,
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	if err := json.Unmarshal(responseBody, target); err != nil {
		return fmt.Errorf("failed to decode registry broker response: %w", err)
	}

	return nil
}
