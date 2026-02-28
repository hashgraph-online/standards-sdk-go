# Contributing to Hashgraph Online HCS SDK (Go)

Thank you for your interest in contributing to the Go SDK for the Hiero Consensus Specifications (HCS).

## Issues and Requests

- Bug reports: [GitHub Issues](https://github.com/hashgraph-online/standards-sdk-go/issues)
- Feature requests: [GitHub Issues](https://github.com/hashgraph-online/standards-sdk-go/issues)

Before opening a new issue, search existing issues first to avoid duplicates.

## Code Contributions

Contributions are accepted through pull requests to `main`.

Please ensure:

1. Your change is scoped and documented.
2. Tests are included or updated for behavior changes.
3. Local validation passes before opening a PR.
4. All commits include DCO sign-off (`git commit -s`).

All contributions are released under [Apache-2.0](./LICENSE).

## Development Setup

```bash
# Clone your fork
git clone https://github.com/<your-username>/standards-sdk-go.git
cd standards-sdk-go

# Fetch dependencies and run tests
go mod tidy
go test ./...
```

For integration tests, configure required environment values locally and run the specific package integration suites you are changing.

## Pull Request Checklist

- [ ] Tests added/updated
- [ ] `go test ./...` passes
- [ ] Public APIs and README are updated when needed
- [ ] PR description explains purpose and scope
- [ ] Commits are signed off (DCO)

## Commit and PR Conventions

Use Conventional Commits:

- `feat:` new functionality
- `fix:` bug fixes
- `docs:` documentation updates
- `refactor:` behavior-preserving refactors
- `test:` test changes
- `chore:` maintenance

## Code of Conduct

By participating, you agree to follow the repository [Code of Conduct](./CODE_OF_CONDUCT.md).
