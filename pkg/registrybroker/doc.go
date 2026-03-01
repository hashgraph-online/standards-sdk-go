// Package registrybroker provides a full client for the HOL
// Registry Broker API. It supports agent search and discovery, adapter
// management, credit operations, verification, ledger-based authentication,
// encrypted chat, feedback, and skill management.
//
// The Registry Broker is the unified discovery and interaction layer for
// AI agents, MCP servers, and decentralized services registered on the
// Hedera public ledger. It aggregates multiple registries (ANS, A2A, MCP)
// and provides trust scores anchored to Hedera consensus.
//
// # Getting Started
//
// Create a Registry Broker client:
//
//	client, err := registrybroker.NewRegistryBrokerClient(
//		registrybroker.RegistryBrokerClientOptions{
//			APIKey:  "<registry-broker-api-key>",
//			BaseURL: "https://hol.org/registry/api/v1",
//		},
//	)
//
//	stats, err := client.Stats(ctx)
//
// # Registry Broker
//
// Learn more about the Registry Broker: https://hol.org/registry
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// This package is part of the HOL Standards SDK for Go.
// See https://hol.org for more information about the HOL ecosystem.
package registrybroker
