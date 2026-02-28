# Hashgraph Online HCS SDK (Go)

[![Go Reference](https://pkg.go.dev/badge/github.com/hashgraph-online/standards-sdk-go.svg)](https://pkg.go.dev/github.com/hashgraph-online/standards-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/hashgraph-online/standards-sdk-go)](https://goreportcard.com/report/github.com/hashgraph-online/standards-sdk-go)
[![Go CI](https://github.com/hashgraph-online/standards-sdk-go/actions/workflows/go-module-release.yml/badge.svg)](https://github.com/hashgraph-online/standards-sdk-go/actions/workflows/go-module-release.yml)
[![GitHub Release](https://img.shields.io/github/v/release/hashgraph-online/standards-sdk-go)](https://github.com/hashgraph-online/standards-sdk-go/releases)
[![License](https://img.shields.io/github/license/hashgraph-online/standards-sdk-go)](./LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/hashgraph-online/standards-sdk-go?style=social)](https://github.com/hashgraph-online/standards-sdk-go/stargazers)
[![CodeSandbox Examples](https://img.shields.io/badge/CodeSandbox-Examples-151515?logo=codesandbox&logoColor=white)](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples)
[![HOL SDK Docs](https://img.shields.io/badge/📚_SDK_Docs-hol.org-4A90D9)](https://hol.org/docs/libraries/standards-sdk/)
[![HCS Standards](https://img.shields.io/badge/📖_HCS_Standards-hol.org-8B5CF6)](https://hol.org/docs/standards)

| ![](./Hashgraph-Online.png) | Go reference implementation of the Hiero Consensus Specifications (HCS) and Registry Broker utilities.<br><br>[📚 Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)<br>[📖 Hiero Consensus Specifications Documentation](https://hol.org/docs/standards) |
| :-------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |

## Quick Start

```bash
cd standards-sdk-go
go mod tidy
go test ./...
```

## Install

```bash
go get github.com/hashgraph-online/standards-sdk-go@latest
```

## CodeSandbox Examples

- [Examples index](./examples/README.md)
- [HCS-2 create registry](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs2-create-registry)
- [HCS-5 build mint transaction](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs5-build-mint)
- [HCS-11 build agent profile](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs11-build-agent-profile)
- [HCS-14 parse UAID](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs14-parse-uaid)
- [HCS-15 build account transaction](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs15-build-account-tx)
- [HCS-16 build flora topic transaction](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs16-build-flora-topic-tx)
- [HCS-17 build state hash message](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs17-build-state-message)
- [HCS-27 publish checkpoint](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/hcs27-publish-checkpoint)
- [Inscriber authenticate + client](https://codesandbox.io/s/github/hashgraph-online/standards-sdk-go/tree/main/examples/inscriber-auth-client)

## Supported Packages

- `pkg/hcs2`: HCS-2 registry topic creation, tx builders, indexed entry operations, memo helpers, mirror reads.
- `pkg/hcs5`: HCS-5 Hashinal minting helpers and end-to-end inscribe+mint workflow.
- `pkg/hcs11`: HCS-11 profile models/builders, validation, inscription, account memo updates, and profile resolution.
- `pkg/hcs14`: HCS-14 UAID generation/parsing plus profile resolution (`_uaid`, `_agent`, ANS `_ans`, and `uaid:did` base DID reconstruction).
- `pkg/hcs15`: HCS-15 base/petal account creation, tx builders, and petal/base key verification helpers.
- `pkg/hcs16`: HCS-16 flora account + topic management, message builders/senders, and threshold-member key assembly helpers.
- `pkg/hcs17`: HCS-17 state-hash topic/message support, deterministic state hash calculators, and verification helpers.
- `pkg/hcs27`: HCS-27 checkpoint topic creation, publish/retrieval, validation, Merkle/proof helpers.
- `pkg/inscriber`: Kiloscribe auth flow, websocket-first high-level inscription utilities, quote generation, bulk-files support, registry-broker quote/job helpers, and skill inscription helpers.
- `pkg/registrybroker`: Full Registry Broker client (search, adapters, agents, credits, verification, ledger auth, chat/encryption, feedback, skills).
- `pkg/mirror`: Mirror node client used by HCS and inscriber packages.
- `pkg/shared`: Network normalization, operator env loading, Hedera client/key parsing helpers.

## Usage Examples

### HCS-2

```go
client, _ := hcs2.NewClient(hcs2.ClientConfig{
	OperatorAccountID:  "0.0.1234",
	OperatorPrivateKey: "<private-key>",
	Network:            "testnet",
})

_, _ = client.CreateRegistry(context.Background(), hcs2.CreateRegistryOptions{
	RegistryType:        hcs2.RegistryTypeIndexed,
	TTL:                 86400,
	UseOperatorAsAdmin:  true,
	UseOperatorAsSubmit: true,
})
```

### HCS-27

```go
client, _ := hcs27.NewClient(hcs27.ClientConfig{
	OperatorAccountID:  "0.0.1234",
	OperatorPrivateKey: "<private-key>",
	Network:            "testnet",
})

metadata := hcs27.CheckpointMetadata{
	Type:   "ans-checkpoint-v1",
	Stream: hcs27.StreamID{Registry: "ans", LogID: "default"},
	Root:   hcs27.RootCommitment{TreeSize: 1, RootHashB64: "<base64url-root>"},
	BatchRange: hcs27.BatchRange{
		Start: 1,
		End:   1,
	},
}
```

### HCS-14

```go
client := hcs14.NewClient(hcs14.ClientOptions{})

result, _ := client.Resolve(
	context.Background(),
	"uaid:aid:ans-godaddy-ote;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1",
)
```

### Inscriber

```go
authClient := inscriber.NewAuthClient("")
authResult, _ := authClient.Authenticate(ctx, accountID, privateKey, inscriber.NetworkTestnet)

client, _ := inscriber.NewClient(inscriber.Config{
	APIKey:  authResult.APIKey,
	Network: inscriber.NetworkTestnet,
})
```

### Registry Broker

```go
client, _ := registrybroker.NewRegistryBrokerClient(registrybroker.RegistryBrokerClientOptions{
	APIKey:  "<registry-broker-api-key>",
	BaseURL: "https://hol.org/registry/api/v1",
})

_, _ = client.Stats(context.Background())
```

## Environment Variables

Common:

- `HEDERA_ACCOUNT_ID`
- `HEDERA_PRIVATE_KEY`
- `HEDERA_NETWORK`
- aliases also supported: `HEDERA_OPERATOR_ID`/`HEDERA_OPERATOR_KEY`, `OPERATOR_ID`/`OPERATOR_KEY`, `ACCOUNT_ID`/`PRIVATE_KEY`

Network-scoped overrides (`pkg/shared`):

- `TESTNET_HEDERA_ACCOUNT_ID`
- `TESTNET_HEDERA_PRIVATE_KEY`
- `MAINNET_HEDERA_ACCOUNT_ID`
- `MAINNET_HEDERA_PRIVATE_KEY`
- aliases also supported: `TESTNET_HEDERA_OPERATOR_ID`/`TESTNET_HEDERA_OPERATOR_KEY`, `MAINNET_HEDERA_OPERATOR_ID`/`MAINNET_HEDERA_OPERATOR_KEY`

The SDK auto-loads `.env` from the current working directory or ancestor directories before resolving credentials.

Inscriber integration:

- `RUN_INTEGRATION=1`
- `RUN_INSCRIBER_INTEGRATION=1`
- `INSCRIPTION_AUTH_BASE_URL` (optional)
- `INSCRIPTION_API_BASE_URL` (optional)
- `INSCRIBER_HEDERA_NETWORK` (optional; defaults to `testnet`)

Registry Broker integration:

- `RUN_INTEGRATION=1`
- `RUN_REGISTRY_BROKER_INTEGRATION=1`
- `REGISTRY_BROKER_API_KEY`
- `REGISTRY_BROKER_BASE_URL` (optional)

## Tests

All packages:

```bash
go test ./...
```

Live HCS + Inscriber integration (no mocks):

```bash
RUN_INTEGRATION=1 \
RUN_INSCRIBER_INTEGRATION=1 \
go test -v ./pkg/hcs2 ./pkg/hcs27 ./pkg/inscriber
```

Live HCS-15 integration (base/petal account flow):

```bash
RUN_INTEGRATION=1 \
go test -v ./pkg/hcs15 -run TestHCS15Integration_CreateBaseAndPetalAccounts
```

Live HCS-16 integration (flora + topic/message flow):

```bash
RUN_INTEGRATION=1 \
go test -v ./pkg/hcs16 -run TestHCS16Integration_CreateFloraAndPublishMessages
```

Live HCS-17 integration (compute + publish state hash):

```bash
RUN_INTEGRATION=1 \
go test -v ./pkg/hcs17 -run TestHCS17Integration_ComputeAndPublishStateHash
```

Live high-level inscriber utilities (websocket default + bulk-files quote):

```bash
RUN_INTEGRATION=1 \
RUN_INSCRIBER_INTEGRATION=1 \
go test -v ./pkg/inscriber -run 'TestInscriberIntegration_(HighLevelInscribe_DefaultWebSocket|GenerateQuote_BulkFiles)'
```

Live Registry Broker skill inscription utility:

```bash
RUN_INTEGRATION=1 \
RUN_REGISTRY_BROKER_INTEGRATION=1 \
REGISTRY_BROKER_API_KEY=<api-key> \
go test -v ./pkg/inscriber -run TestInscriberIntegration_RegistryBrokerSkillInscribe
```

Live HCS-5 mint integration (requires target NFT token and supply key):

```bash
RUN_INTEGRATION=1 \
HCS5_INTEGRATION_TOKEN_ID=<token-id> \
HCS5_INTEGRATION_SUPPLY_KEY=<private-key> \
go test -v ./pkg/hcs5 -run TestHCS5Integration_MintWithExistingHCS1Topic
```

Live HCS-11 profile lookup integration:

```bash
RUN_INTEGRATION=1 \
HCS11_INTEGRATION_ACCOUNT_ID=<account-id> \
HCS11_INTEGRATION_NETWORK=testnet \
go test -v ./pkg/hcs11 -run TestHCS11Integration_FetchProfileByAccountID
```

Live HCS-14 integration (DNS/Web resolution):

```bash
RUN_INTEGRATION=1 \
go test -v ./pkg/hcs14 -run TestHCS14Integration_ANSDNSWebResolution
```

Live Registry Broker integration:

```bash
RUN_INTEGRATION=1 \
RUN_REGISTRY_BROKER_INTEGRATION=1 \
REGISTRY_BROKER_API_KEY=<api-key> \
go test -v ./pkg/registrybroker
```

## Contributing

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) before contributing.

## Security

For security concerns, see [SECURITY.md](./SECURITY.md).

## Maintainers

See [MAINTAINERS.md](./MAINTAINERS.md).

## Resources

- [Hiero Consensus Specifications (HCS) Documentation](https://hol.org/docs/standards)
- [Hedera Documentation](https://docs.hedera.com)
- [Telegram Community](https://t.me/hashinals)

## License

Apache-2.0
