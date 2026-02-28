// Package hcs2 implements the HCS-2 Topic Registry specification for the
// Hedera Consensus Service (HCS). It provides registry topic creation,
// transaction builders, indexed entry operations, memo helpers, and
// mirror-node reads for managing on-chain topic registries.
//
// HCS-2 defines a standard for creating and managing topic-based registries
// on HCS, enabling decentralized, append-only data structures anchored to
// the Hedera public ledger.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-2
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Getting Started
//
// Create a registry client and initialize an indexed registry:
//
//	client, err := hcs2.NewClient(hcs2.ClientConfig{
//		OperatorAccountID:  "0.0.1234",
//		OperatorPrivateKey: "<private-key>",
//		Network:            "testnet",
//	})
//
//	result, err := client.CreateRegistry(ctx, hcs2.CreateRegistryOptions{
//		RegistryType:        hcs2.RegistryTypeIndexed,
//		TTL:                 86400,
//		UseOperatorAsAdmin:  true,
//		UseOperatorAsSubmit: true,
//	})
//
// This package is part of the Hashgraph Online Standards SDK for Go.
// See https://hol.org for more information about the Hashgraph Online ecosystem.
package hcs2
