# Registry Broker Skill Domain Proof

This example runs the full DNS TXT domain proof flow for a published skill:

1. Ledger-authenticates against Registry Broker.
2. Creates a DNS TXT challenge for `SKILL_NAME`.
3. Optionally upserts the TXT record in Cloudflare.
4. Verifies the domain proof.
5. Prints trust metric before/after, including `verified.domainProof`.

## Required environment variables

```bash
export SKILL_NAME="your-skill-name"
export TESTNET_HEDERA_ACCOUNT_ID="0.0.xxxxx"
export TESTNET_HEDERA_PRIVATE_KEY="302e..."
```

Optional:

```bash
export REGISTRY_BROKER_BASE_URL="https://registry-staging.hol.org/api/v1"
export SKILL_VERSION="1.0.0"
export SKILL_DOMAIN_PROOF_DOMAIN="example.com"
export SKILL_DOMAIN_PROOF_WAIT_DNS_SECONDS="180"
export SKILL_DOMAIN_PROOF_AUTO_DNS="true"
export CLOUDFLARE_API_TOKEN="..."
export CLOUDFLARE_ZONE_ID="..."
```

## Run

```bash
go run ./examples/registry-broker-skill-domain-proof
```

