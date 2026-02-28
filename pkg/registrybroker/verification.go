package registrybroker

import (
	"context"
	"net/http"
	"strings"
)

func (c *RegistryBrokerClient) GetVerificationStatus(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/verification/status/" + percentPath(uaid)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

func (c *RegistryBrokerClient) CreateVerificationChallenge(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	body := JSONObject{"uaid": strings.TrimSpace(uaid)}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/verification/challenge",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) GetVerificationChallenge(
	ctx context.Context,
	challengeID string,
) (JSONObject, error) {
	if err := ensureNonEmpty(challengeID, "challengeId"); err != nil {
		return nil, err
	}
	path := "/verification/challenge/" + percentPath(challengeID)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

func (c *RegistryBrokerClient) VerifyVerificationChallenge(
	ctx context.Context,
	payload VerifyVerificationChallengeRequest,
) (JSONObject, error) {
	body := JSONObject{
		"challengeId": payload["challengeId"],
	}
	if methodValue, ok := payload["method"]; ok && methodValue != nil {
		body["method"] = methodValue
	} else {
		body["method"] = "moltbook-post"
	}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/verification/verify",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) GetVerificationOwnership(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/verification/ownership/" + percentPath(uaid)
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}

func (c *RegistryBrokerClient) VerifySenderOwnership(
	ctx context.Context,
	uaid string,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	body := JSONObject{"uaid": strings.TrimSpace(uaid)}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/verification/verify-sender",
		body,
		map[string]string{"content-type": "application/json"},
	)
}
