// Package hcs27 implements the HCS-27 Checkpoint specification for the Hedera
// Consensus Service (HCS). It provides checkpoint topic creation, publish and
// retrieval operations, validation, and Merkle tree and proof helpers for
// anchoring verifiable data checkpoints to the Hedera public ledger.
//
// HCS-27 defines a standard for publishing and verifying Merkle-root-based
// checkpoints on HCS, enabling applications to create tamper-evident,
// consensus-anchored proofs of data integrity with efficient inclusion proofs.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-27
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Publishing a Checkpoint
//
// Create a client and publish a checkpoint with Merkle root commitment:
//
//	client, err := hcs27.NewClient(hcs27.ClientConfig{
//		OperatorAccountID:  "0.0.1234",
//		OperatorPrivateKey: "<private-key>",
//		Network:            "testnet",
//	})
//
//	metadata := hcs27.CheckpointMetadata{
//		Type:   "ans-checkpoint-v1",
//		Stream: hcs27.StreamID{Registry: "ans", LogID: "default"},
//		Root:   hcs27.RootCommitment{TreeSize: 100, RootHashB64: "<root>"},
//	}
//
// This package is part of the HOL Standards SDK for Go.
// See https://hol.org for more information about the HOL ecosystem.
package hcs27
