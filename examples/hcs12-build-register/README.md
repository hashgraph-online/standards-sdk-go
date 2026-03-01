# HCS-12: Build a Register Payload

This example builds an [HCS-12](https://hol.org/docs/standards/hcs-12) register payload transaction.

## What it does

1. Constructs an HCS-12 registration payload for the Registry Broker.
2. Demonstrates the agent registration workflow.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/hcs12-build-register
```

## Learn More

- [HCS-12 Specification](https://hol.org/docs/standards/hcs-12)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Registry Broker](https://hol.org/registry)
- [Hashgraph Online](https://hol.org)
