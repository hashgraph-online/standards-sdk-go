# Hashgraph Online LangChainGo Tools

This package provides a suite of tools for the [LangChainGo](https://github.com/tmc/langchaingo) AI agent framework, enabling AI agents to interact directly with the [Hiero Consensus Specifications](https://hol.org/docs/standards) on the Hedera public ledger.

## Installation

```bash
go get github.com/hashgraph-online/standards-sdk-go/adapters/langchaingo
```

## Available Tools

### UAID Resolver Tool (`langchaingo.UAIDResolverTool`)

Enables an AI agent to look up [Universal Agent IDs (HCS-14)](https://hol.org/docs/standards/hcs-14) securely via the Hedera network, resolving them into their associated JSON profiles (which may include the agent's endpoint, authentication public keys, and metadata).

**Usage with an Agent:**

```go
package main

import (
"context"
"fmt"

"github.com/tmc/langchaingo/llms/openai"
"github.com/tmc/langchaingo/agents"
"github.com/tmc/langchaingo/tools"

hol_langchaingo "github.com/hashgraph-online/standards-sdk-go/adapters/langchaingo"
)

func main() {
llm, _ := openai.New()

// Initialize the UAID resolving tool
uaidTool := hol_langchaingo.NewUAIDResolverTool(nil)

// Combine with a Web Search tool or others
agentTools := []tools.Tool{uaidTool}

executor, _ := agents.Initialize(
llm,
agentTools,
agents.ZeroShotReactDescription,
)

response, _ := agents.Run(context.Background(), executor, "Find the endpoint for the agent with UAID: uaid:aid:my-agent;registry=ans;proto=a2a")
fmt.Println(response)
}
```

## Learn More
- [Hashgraph Online](https://hol.org)
- [Standards SDK Documentation](https://hol.org/docs/libraries/standards-sdk/)
- [HCS Specifications](https://hol.org/docs/standards)
