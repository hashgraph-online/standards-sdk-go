// Package hcs17 implements the HCS-17 State Hash specification for the Hedera
// Consensus Service (HCS). It provides state-hash topic and message support,
// deterministic state hash calculators, and verification helpers for anchoring
// application state to the Hedera public ledger.
//
// HCS-17 defines a standard for computing, publishing, and verifying
// deterministic state hashes, enabling applications to prove their state
// integrity against an immutable, consensus-ordered record on HCS.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-17
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// This package is part of the HOL Standards SDK for Go.
// See https://hol.org for more information about the HOL ecosystem.
package hcs17
