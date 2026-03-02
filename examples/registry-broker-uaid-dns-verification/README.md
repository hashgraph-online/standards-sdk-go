# Registry Broker UAID DNS Verification

This example proves the UAID DNS TXT verification flow end to end with the Go `RegistryBrokerClient`:

1. Ledger-authenticate against Registry Broker.
2. Call `VerifyUaidDNSTXT` for a target UAID.
3. Read stored status with `GetVerificationDNSStatus(..., refresh=false)`.
4. Force a live refresh with `GetVerificationDNSStatus(..., refresh=true)`.

## Required environment variables

Use either API key auth or ledger auth.

Option A (API key):

```bash
export REGISTRY_BROKER_API_KEY="rbk_..."
export UAID_DNS_DEMO_UAID="uaid:aid:...;uid=...;proto=a2a;nativeId=agent.hol.org"
```

Option B (ledger):

```bash
export UAID_DNS_DEMO_UAID="uaid:aid:...;uid=...;proto=a2a;nativeId=agent.hol.org"
export TESTNET_HEDERA_ACCOUNT_ID="0.0.xxxxx"
export TESTNET_HEDERA_PRIVATE_KEY="302e..."
```

Optional:

```bash
export REGISTRY_BROKER_BASE_URL="https://hol.org/registry/api/v1"
export UAID_DNS_DEMO_PERSIST="true"
export LEDGER_NETWORK="testnet"
```

## Run

```bash
go run ./examples/registry-broker-uaid-dns-verification --uaid="uaid:aid:...;uid=...;proto=a2a;nativeId=agent.hol.org"
```
