# HCS-17: Build a State Hash Message

This example builds an [HCS-17](https://hol.org/docs/standards/hcs-17) state hash message transaction.

## What it does

1. Constructs an HCS-17 state hash message for anchoring application state.
2. Demonstrates deterministic state-hash publication on HCS.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/hcs17-build-state-message
```

## Learn More

- [HCS-17 Specification](https://hol.org/docs/standards/hcs-17)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Hashgraph Online](https://hol.org)
