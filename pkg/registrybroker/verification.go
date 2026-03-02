package registrybroker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetVerificationStatus returns the requested value.
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

// CreateVerificationChallenge creates the requested resource.
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

// GetVerificationChallenge returns the requested value.
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

// VerifyVerificationChallenge performs the requested operation.
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

// GetVerificationOwnership returns the requested value.
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

// VerifySenderOwnership performs the requested operation.
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

// VerifyUaidDnsTXT performs DNS TXT verification for the provided UAID.
func (c *RegistryBrokerClient) VerifyUaidDnsTXT(
	ctx context.Context,
	payload VerificationDnsVerifyRequest,
) (JSONObject, error) {
	if err := ensureNonEmpty(payload.UAID, "uaid"); err != nil {
		return nil, err
	}
	body := JSONObject{
		"uaid": strings.TrimSpace(payload.UAID),
	}
	if payload.Persist != nil {
		body["persist"] = *payload.Persist
	}
	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/verification/dns/verify",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

// GetVerificationDNSStatus returns stored or live DNS TXT verification status for a UAID.
func (c *RegistryBrokerClient) GetVerificationDNSStatus(
	ctx context.Context,
	uaid string,
	query VerificationDnsStatusQuery,
) (JSONObject, error) {
	if err := ensureNonEmpty(uaid, "uaid"); err != nil {
		return nil, err
	}
	path := "/verification/dns/status/" + percentPath(uaid)
	params := url.Values{}
	if query.Refresh != nil {
		params.Set("refresh", strconv.FormatBool(*query.Refresh))
	}
	if query.Persist != nil {
		params.Set("persist", strconv.FormatBool(*query.Persist))
	}
	if encoded := params.Encode(); encoded != "" {
		path = path + "?" + encoded
	}
	return c.requestJSON(ctx, http.MethodGet, path, nil, nil)
}
