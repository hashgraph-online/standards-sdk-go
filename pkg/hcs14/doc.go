// Package hcs14 implements the HCS-14 Universal Agent ID (UAID) specification
// for the Hedera Consensus Service (HCS). It provides UAID generation, parsing,
// profile resolution via _uaid, _agent, and ANS _ans DNS TXT records, and
// uaid:did base DID reconstruction.
//
// HCS-14 defines a universal, verifiable identity standard for AI agents,
// enabling cross-platform agent discovery and authentication anchored to the
// Hedera public ledger.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-14
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Resolving a UAID
//
// Use the client to resolve a UAID string to its associated profile:
//
//	client := hcs14.NewClient(hcs14.ClientOptions{})
//	result, err := client.Resolve(ctx,
//		"uaid:aid:my-agent;registry=ans;proto=a2a",
//	)
//
// This package is part of the HOL Standards SDK for Go.
// See https://hol.org for more information about the HOL ecosystem.
package hcs14
