// Package hcs15 implements the HCS-15 Base/Petal Account specification for the
// Hedera Consensus Service (HCS). It provides base and petal account creation,
// transaction builders, and key verification helpers for hierarchical account
// structures on the Hedera public ledger.
//
// HCS-15 defines a standard for creating hierarchical account relationships
// where a base account can spawn and manage petal accounts, enabling structured
// identity and permission models for decentralized applications.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-15
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// This package is part of the Hashgraph Online Standards SDK for Go.
// See https://hol.org for more information about the Hashgraph Online ecosystem.
package hcs15
