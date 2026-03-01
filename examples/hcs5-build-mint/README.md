# HCS-5: Build a Hashinal Mint Transaction

This example builds an [HCS-5 Hashinal](https://hol.org/docs/standards/hcs-5) mint transaction with HCS-1 HRL metadata.

## What it does

1. Constructs an HCS-5 mint transaction referencing on-chain HCS data.
2. Demonstrates the Hashinal NFT minting workflow.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/hcs5-build-mint
```

## Learn More

- [HCS-5 Specification](https://hol.org/docs/standards/hcs-5)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Hashgraph Online](https://hol.org)
