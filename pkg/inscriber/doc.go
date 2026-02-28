// Package inscriber provides the Kiloscribe authentication flow and high-level
// inscription utilities for the Hashgraph Online ecosystem. It supports
// websocket-first inscription, quote generation, bulk-files support,
// Registry Broker quote and job helpers, and skill inscription helpers.
//
// The inscriber package is the primary interface for writing data to the Hedera
// public ledger via the Kiloscribe inscription service. It handles
// authentication, file chunking, websocket transport, and cost estimation.
//
// # Authentication
//
// Authenticate with the inscription service using a Hedera account:
//
//	authClient := inscriber.NewAuthClient("")
//	result, err := authClient.Authenticate(ctx, accountID, privateKey, inscriber.NetworkTestnet)
//
//	client, err := inscriber.NewClient(inscriber.Config{
//		APIKey:  result.APIKey,
//		Network: inscriber.NetworkTestnet,
//	})
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Inscription Service
//
// Learn more about inscriptions: https://hol.org/docs/standards/hcs-1
//
// This package is part of the Hashgraph Online Standards SDK for Go.
// See https://hol.org for more information about the Hashgraph Online ecosystem.
package inscriber
