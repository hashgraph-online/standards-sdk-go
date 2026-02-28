package registrybroker

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
)

func (c *RegistryBrokerClient) PurchaseCreditsWithHbar(
	ctx context.Context,
	params PurchaseCreditsWithHbarParams,
) (JSONObject, error) {
	if err := ensureNonEmpty(params.AccountID, "accountId"); err != nil {
		return nil, err
	}
	if err := ensureNonEmpty(params.PrivateKey, "privateKey"); err != nil {
		return nil, err
	}
	if params.HbarAmount <= 0 {
		return nil, fmt.Errorf("hbarAmount must be a positive number")
	}

	body := JSONObject{
		"accountId":  params.AccountID,
		"payerKey":   params.PrivateKey,
		"hbarAmount": calculateHbarAmountParam(params.HbarAmount),
	}
	if strings.TrimSpace(params.Memo) != "" {
		body["memo"] = strings.TrimSpace(params.Memo)
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/credits/purchase",
		body,
		map[string]string{"content-type": "application/json"},
	)
}

func (c *RegistryBrokerClient) GetX402Minimums(ctx context.Context) (JSONObject, error) {
	return c.requestJSON(ctx, http.MethodGet, "/credits/purchase/x402/minimums", nil, nil)
}

func (c *RegistryBrokerClient) PurchaseCreditsWithX402(
	ctx context.Context,
	params PurchaseCreditsWithX402Params,
) (JSONObject, error) {
	if err := ensureNonEmpty(params.AccountID, "accountId"); err != nil {
		return nil, err
	}
	if params.Credits <= 0 {
		return nil, fmt.Errorf("credits must be a positive number")
	}
	if params.USDAmount != nil && *params.USDAmount <= 0 {
		return nil, fmt.Errorf("usdAmount must be a positive number")
	}

	body := JSONObject{
		"accountId": params.AccountID,
		"credits":   params.Credits,
	}
	if params.USDAmount != nil {
		body["usdAmount"] = *params.USDAmount
	}
	if strings.TrimSpace(params.Description) != "" {
		body["description"] = strings.TrimSpace(params.Description)
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	headers := map[string]string{"content-type": "application/json"}
	for key, value := range params.PaymentHeaders {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}

	return c.requestJSON(
		ctx,
		http.MethodPost,
		"/credits/purchase/x402",
		body,
		headers,
	)
}

func (c *RegistryBrokerClient) BuyCreditsWithX402(
	ctx context.Context,
	params BuyCreditsWithX402Params,
) (JSONObject, error) {
	purchase := PurchaseCreditsWithX402Params{
		AccountID:   params.AccountID,
		Credits:     params.Credits,
		USDAmount:   params.USDAmount,
		Description: params.Description,
		Metadata:    params.Metadata,
	}

	if strings.TrimSpace(params.EVMPrivateKey) != "" {
		purchase.PaymentHeaders = map[string]string{
			"x-evm-private-key": strings.TrimSpace(params.EVMPrivateKey),
		}
	}
	if strings.TrimSpace(params.Network) != "" {
		if purchase.PaymentHeaders == nil {
			purchase.PaymentHeaders = map[string]string{}
		}
		purchase.PaymentHeaders["x-x402-network"] = strings.TrimSpace(params.Network)
	}
	if strings.TrimSpace(params.RPCURL) != "" {
		if purchase.PaymentHeaders == nil {
			purchase.PaymentHeaders = map[string]string{}
		}
		purchase.PaymentHeaders["x-x402-rpc-url"] = strings.TrimSpace(params.RPCURL)
	}

	return c.PurchaseCreditsWithX402(ctx, purchase)
}

func calculateHbarAmountParam(value float64) float64 {
	tinybars := math.Ceil(value * 1e8)
	if tinybars <= 0 {
		return value
	}
	return tinybars / 1e8
}
