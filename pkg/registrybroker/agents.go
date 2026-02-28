package registrybroker

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// ResolveUaid resolves the requested identifier data.
func (c *RegistryBrokerClient) ResolveUaid(ctx context.Context, uaid string) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/resolve/" + percentPath(uaid)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// PerformRegisterAgent performs the requested operation.
func (c *RegistryBrokerClient) PerformRegisterAgent(
	ctx context.Context,
	payload AgentRegistrationRequest,
) (JSONObject, error) {
	body := serialiseAgentRegistrationRequest(payload)
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/register",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// RegisterAgent registers the requested resource.
func (c *RegistryBrokerClient) RegisterAgent(
	ctx context.Context,
	payload AgentRegistrationRequest,
	options *RegisterAgentOptions,
) (JSONObject, error) {
	var autoTop *AutoTopUpOptions
	if options != nil && options.AutoTopUp != nil {
		autoTop = options.AutoTopUp
	} else {
		autoTop = c.registrationAutoTop
	}

	if autoTop == nil {
		return c.PerformRegisterAgent(ctx, payload)
	}

	if err := c.ensureCreditsForRegistration(ctx, payload, autoTop); err != nil {
		return nil, err
	}

	retried := false
	for {
		result, err := c.PerformRegisterAgent(ctx, payload)
		if err == nil {
			return result, nil
		}
		shortfall, hasShortfall := c.extractInsufficientCreditsDetails(err)
		if hasShortfall && !retried && shortfall > 0 {
			if topUpErr := c.ensureCreditsForRegistration(ctx, payload, autoTop); topUpErr != nil {
				return nil, topUpErr
			}
			retried = true
			continue
		}
		return nil, err
	}
}

// GetRegistrationQuote returns the requested value.
func (c *RegistryBrokerClient) GetRegistrationQuote(
	ctx context.Context,
	payload AgentRegistrationRequest,
) (JSONObject, error) {
	body := serialiseAgentRegistrationRequest(payload)
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/register/quote",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// UpdateAgent updates the requested resource.
func (c *RegistryBrokerClient) UpdateAgent(
	ctx context.Context,
	uaid string,
	payload AgentRegistrationRequest,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	body := serialiseAgentRegistrationRequest(payload)
	path := "/register/" + percentPath(uaid)
	return c.requestJSON(
		ctx,
		http.MethodPut,
		path,
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// GetRegisterStatus returns the requested value.
func (c *RegistryBrokerClient) GetRegisterStatus(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/register/status/" + percentPath(uaid)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// RegisterOwnedMoltbookAgent registers the requested resource.
func (c *RegistryBrokerClient) RegisterOwnedMoltbookAgent(
	ctx context.Context,
	uaid string,
	payload MoltbookOwnerRegistrationUpdateRequest,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/register/" + percentPath(uaid)
	return c.requestJSON(
		ctx,
		http.MethodPut,
		path,
		payload,
		map[string]string{"content-type": "application/json"},
	)
}

// GetRegistrationProgress returns the requested value.
func (c *RegistryBrokerClient) GetRegistrationProgress(
	ctx context.Context,
	attemptID string,
) (JSONObject, error) {
	if err := ensureNonEmpty(attemptID, "attemptID"); err != nil {
		return nil, err
	}
	path := "/register/progress/" + percentPath(attemptID)
	result, err := c.requestJSON(ctx, http.MethodGet, path, nil, nil)
	if err == nil {
		if progressValue, exists := result["progress"]; exists {
			if progressMap, ok := progressValue.(map[string]any); ok {
				return progressMap, nil
			}
			if progressMap, ok := progressValue.(JSONObject); ok {
				return progressMap, nil
			}
		}
		return nil, nil
	}
	brokerErr, ok := err.(*RegistryBrokerError)
	if ok && brokerErr.Status == http.StatusNotFound {
		return nil, nil
	}
	return nil, err
}

// WaitForRegistrationCompletion performs the requested operation.
func (c *RegistryBrokerClient) WaitForRegistrationCompletion(
	ctx context.Context,
	attemptID string,
	options RegistrationProgressWaitOptions,
) (JSONObject, error) {
	if err := ensureNonEmpty(attemptID, "attemptID"); err != nil {
		return nil, err
	}

	interval := options.Interval
	if interval <= 0 {
		interval = DefaultProgressInterval
	}

	timeout := options.Timeout
	if timeout <= 0 {
		timeout = DefaultProgressTimeout
	}

	throwOnFailure := true
	if options.ThrowOnFailure != nil {
		throwOnFailure = *options.ThrowOnFailure
	}

	started := time.Now()
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		progress, err := c.GetRegistrationProgress(ctx, attemptID)
		if err != nil {
			return nil, err
		}
		if progress != nil {
			if options.OnProgress != nil {
				options.OnProgress(progress)
			}
			status := strings.ToLower(strings.TrimSpace(stringField(progress, "status")))
			if status == "completed" {
				return progress, nil
			}
			if status == "partial" || status == "failed" {
				if throwOnFailure {
					return nil, &RegistryBrokerError{
						Message:    "registration did not complete successfully",
						Status:     http.StatusConflict,
						StatusText: status,
						Body:       progress,
					}
				}
				return progress, nil
			}
		}

		if time.Since(started) >= timeout {
			return nil, fmt.Errorf("registration progress polling timed out after %s", timeout.String())
		}

		if err := c.delay(ctx, interval); err != nil {
			return nil, err
		}
	}
}

// ValidateUaid validates the provided input value.
func (c *RegistryBrokerClient) ValidateUaid(ctx context.Context, uaid string) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/uaids/validate/" + percentPath(uaid)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// GetUaidConnectionStatus returns the requested value.
func (c *RegistryBrokerClient) GetUaidConnectionStatus(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/uaids/connections/" + percentPath(uaid) + "/status"
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

// CloseUaidConnection performs the requested operation.
func (c *RegistryBrokerClient) CloseUaidConnection(ctx context.Context, uaid string) error {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return err
	}
	path := "/uaids/connections/" + percentPath(uaid)
	return c.requestNoResponse(ctx, http.MethodDelete, path, nil, nil)
}

// DashboardStats performs the requested operation.
func (c *RegistryBrokerClient) DashboardStats(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/dashboard/stats", nil, nil)
}

func (c *RegistryBrokerClient) ensureCreditsForRegistration(
	ctx context.Context,
	payload AgentRegistrationRequest,
	autoTop *AutoTopUpOptions,
) error {
	if autoTop == nil {
		return nil
	}
	if err := ensureNonEmpty(autoTop.AccountID, "autoTopUp.accountId"); err != nil {
		return err
	}
	if err := ensureNonEmpty(autoTop.PrivateKey, "autoTopUp.privateKey"); err != nil {
		return err
	}

	for attempt := 0; attempt < 3; attempt++ {
		quote, err := c.GetRegistrationQuote(ctx, payload)
		if err != nil {
			return err
		}
		shortfall, _ := getNumberField(quote, "shortfallCredits")
		if shortfall <= 0 {
			return nil
		}
		creditsPerHBAR, ok := getNumberField(quote, "creditsPerHbar")
		if !ok || creditsPerHBAR <= 0 {
			return fmt.Errorf("unable to determine credits per HBAR for auto top-up")
		}
		creditsToPurchase := resolveCreditsToPurchase(shortfall)
		hbarAmount := calculateRegistrationHbarAmount(creditsToPurchase, creditsPerHBAR)
		_, err = c.PurchaseCreditsWithHbar(ctx, PurchaseCreditsWithHbarParams{
			AccountID:  strings.TrimSpace(autoTop.AccountID),
			PrivateKey: strings.TrimSpace(autoTop.PrivateKey),
			HbarAmount: hbarAmount,
			Memo:       autoTop.Memo,
			Metadata: JSONObject{
				"shortfallCredits": shortfall,
				"requiredCredits":  quote["requiredCredits"],
				"purchasedCredits": creditsToPurchase,
			},
		})
		if err != nil {
			return err
		}
	}

	finalQuote, err := c.GetRegistrationQuote(ctx, payload)
	if err != nil {
		return err
	}
	shortfall, _ := getNumberField(finalQuote, "shortfallCredits")
	if shortfall > 0 {
		return fmt.Errorf("unable to purchase sufficient credits for registration")
	}
	return nil
}

func resolveCreditsToPurchase(shortfall float64) float64 {
	if !isFinitePositive(shortfall) {
		return 0
	}
	return math.Max(math.Ceil(shortfall), MinimumRegistrationTopUpCount)
}

func calculateRegistrationHbarAmount(creditsToPurchase float64, creditsPerHBAR float64) float64 {
	if creditsPerHBAR <= 0 || creditsToPurchase <= 0 {
		return 0
	}
	rawHBAR := creditsToPurchase / creditsPerHBAR
	tinybars := math.Ceil(rawHBAR * 1e8)
	return tinybars / 1e8
}

func isFinitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func stringField(source JSONObject, key string) string {
	raw, exists := source[key]
	if !exists || raw == nil {
		return ""
	}
	if typed, ok := raw.(string); ok {
		return typed
	}
	return fmt.Sprint(raw)
}

func serialiseAgentRegistrationRequest(payload AgentRegistrationRequest) JSONObject {
	body := JSONObject{}
	for key, value := range payload {
		if value == nil {
			continue
		}
		body[key] = value
	}
	return body
}
