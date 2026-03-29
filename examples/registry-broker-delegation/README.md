# Registry Broker Delegation

This example calls the broker-native delegation planner and prints the ranked
opportunities plus the top candidate per opportunity.

Optional environment variables:

```bash
export REGISTRY_BROKER_BASE_URL="https://hol.org/registry/api/v1"
export REGISTRY_BROKER_API_KEY="..."
export REGISTRY_BROKER_DELEGATION_TASK="Review an SDK PR and split out docs and verification subtasks."
export REGISTRY_BROKER_DELEGATION_CONTEXT="Need a docs-focused sidecar pass."
export REGISTRY_BROKER_DELEGATION_LIMIT="3"
```

## Run

```bash
go run ./examples/registry-broker-delegation
```
