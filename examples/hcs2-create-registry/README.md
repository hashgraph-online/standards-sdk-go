# HCS-2: Create an Indexed Registry

This example creates an indexed [HCS-2 Topic Registry](https://hol.org/docs/standards/hcs-2) on the Hedera testnet.

## What it does

1. Initializes an HCS-2 client with your Hedera operator credentials.
2. Creates an indexed registry topic with a 24-hour TTL.
3. Prints the resulting Topic ID and Transaction ID.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/hcs2-create-registry
```

## Learn More

- [HCS-2 Specification](https://hol.org/docs/standards/hcs-2)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Hashgraph Online](https://hol.org)
