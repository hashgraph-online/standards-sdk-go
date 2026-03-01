// Package hcs11 implements the HCS-11 Profile Metadata specification for the
// Hedera Consensus Service (HCS). It provides profile models, builders,
// validation, inscription, account memo updates, and profile resolution
// for managing identity profiles on the Hedera public ledger.
//
// HCS-11 defines a standard for creating and resolving structured profile
// metadata (persons, agents, organizations, MCP servers, and flora accounts)
// that are anchored to HCS topics and discoverable via account memos.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-11
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Profile Builders
//
// The package includes specialized builders for different profile types:
//
//   - [AgentBuilder] for AI agent profiles
//   - [PersonBuilder] for individual person profiles
//   - [FloraBuilder] for flora (group/organization) profiles
//   - [MCPServerBuilder] for Model Context Protocol server profiles
//
// This package is part of the HOL Standards SDK for Go.
// See https://hol.org for more information about the HOL ecosystem.
package hcs11
