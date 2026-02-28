// Package hcs20 implements the HCS-20 Auditable Points standard for Hedera
// Consensus Service topics. It provides message validation, transaction
// builders, a submission client, and a mirror-driven indexer for deriving
// point balances and supply state.
//
// # Specification
//
// Full specification: https://hol.org/docs/standards/hcs-20
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Build and Submit HCS-20 Messages
//
// Build deploy/mint/transfer/burn/register payload transactions:
//
//	tx, err := hcs20.BuildHCS20DeployTx(hcs20.DeployTxParams{
//		TopicID: "0.0.12345",
//		Name:    "Loyalty Points",
//		Tick:    "loyal",
//		Max:     "1000000",
//	})
//
// # Client Usage
//
// Create a client and deploy points on testnet:
//
//	client, err := hcs20.NewClient(hcs20.ClientConfig{
//		OperatorAccountID:  "0.0.1234",
//		OperatorPrivateKey: "<private-key>",
//		Network:            "testnet",
//	})
//
//	info, err := client.DeployPoints(context.Background(), hcs20.DeployPointsOptions{
//		Name: "Loyalty Points",
//		Tick: "loyal",
//		Max:  "1000000",
//	})
package hcs20
