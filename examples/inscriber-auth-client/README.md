# Inscriber: Authenticate and Initialize

This example authenticates against [Kiloscribe](https://hol.org/docs/standards/hcs-1) and initializes an inscriber client.

## What it does

1. Authenticates a Hedera account with the Kiloscribe inscription service.
2. Initializes an inscriber client for writing data to HCS topics.

## Run

```bash
export HEDERA_ACCOUNT_ID="0.0.xxxxx"
export HEDERA_PRIVATE_KEY="302..."
go run ./examples/inscriber-auth-client
```

## Learn More

- [HCS-1 Specification](https://hol.org/docs/standards/hcs-1)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [Hashgraph Online](https://hol.org)
