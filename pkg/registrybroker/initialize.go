package registrybroker

import (
	"context"
	"fmt"
	"strings"
)

func InitializeAgent(
	ctx context.Context,
	options InitializeAgentClientOptions,
) (*InitializedAgentClient, error) {
	client, err := NewRegistryBrokerClient(options.RegistryBrokerClientOptions)
	if err != nil {
		return nil, err
	}

	ensureKey := true
	if options.EnsureEncryptionKey != nil {
		ensureKey = *options.EnsureEncryptionKey
	}

	var encryption JSONObject
	if ensureKey {
		uaid := strings.TrimSpace(options.UAID)
		if uaid == "" {
			return nil, fmt.Errorf("uaid is required when ensureEncryptionKey is enabled")
		}
		ensureOptions := EnsureAgentKeyOptions{
			UAID:              uaid,
			GenerateIfMissing: true,
		}
		if options.EnsureEncryptionOptions != nil {
			ensureOptions = *options.EnsureEncryptionOptions
			ensureOptions.UAID = uaid
		}
		encryption, err = client.EnsureAgentKey(ctx, ensureOptions)
		if err != nil {
			return nil, err
		}
	}

	return &InitializedAgentClient{
		Client:     client,
		Encryption: encryption,
	}, nil
}

func IsPendingRegisterAgentResponse(response JSONObject) bool {
	return strings.EqualFold(strings.TrimSpace(stringField(response, "status")), "pending")
}

func IsPartialRegisterAgentResponse(response JSONObject) bool {
	status := strings.EqualFold(strings.TrimSpace(stringField(response, "status")), "partial")
	success, hasSuccess := response["success"].(bool)
	return status && hasSuccess && !success
}

func IsSuccessRegisterAgentResponse(response JSONObject) bool {
	success, hasSuccess := response["success"].(bool)
	if !hasSuccess || !success {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(stringField(response, "status")), "pending")
}
