# HCS-27: Publish a Checkpoint

This example publishes an [HCS-27](https://hol.org/docs/standards/hcs-27) checkpoint message.

## What it does

1. Creates an HCS-27 checkpoint with a Merkle root commitment.
2. Publishes the checkpoint to an HCS topic on testnet.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/hcs27-publish-checkpoint
```

## Learn More

- [HCS-27 Specification](https://hol.org/docs/standards/hcs-27)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Hashgraph Online](https://hol.org)
