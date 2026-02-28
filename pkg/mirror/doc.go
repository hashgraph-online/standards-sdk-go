// Package mirror provides a Hedera Mirror Node client used by the HCS and
// inscriber packages in the Hashgraph Online Standards SDK. It handles
// topic info lookups, message retrieval, and consensus data queries against
// the Hedera mirror node REST API.
//
// The mirror node provides a read-only view of the Hedera public ledger,
// enabling applications to query historical transactions, topic messages,
// and account state without submitting transactions to the network.
//
// # SDK Documentation
//
// SDK documentation and guides: https://hol.org/docs/libraries/standards-sdk/
//
// # Hedera Mirror Node
//
// Learn more about Hedera: https://docs.hedera.com
//
// This package is part of the Hashgraph Online Standards SDK for Go.
// See https://hol.org for more information about the Hashgraph Online ecosystem.
package mirror
