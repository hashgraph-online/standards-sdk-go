// Package langchaingo provides Hashgraph Online tools and integrations for the
// tmc/langchaingo AI agent framework.
//
// These tools allow Langchain-powered agents to interact seamlessly with
// decentralized Hashgraph Online services, such as resolving Universal Agent
// IDs (UAID) to locate and securely communicate with other AI agents.
//
// # Available Tools
//
//   - UAIDResolverTool: Resolves an HCS-14 UAID string into a JSON profile.
//
// # Usage
//
// Add the tool to your agent's toolkit configuration:
//
//	resolverTool := langchaingo.NewUAIDResolverTool(nil)
//	agent := initialize.NewSingleActionAgent(llm, []tools.Tool{resolverTool})
//
//	// The agent can now automatically look up an identity when it sees a UAID.
//
// # Documentation
//
// SDK documentation: https://hol.org/docs/libraries/standards-sdk/
//
// Langchaingo: https://github.com/tmc/langchaingo
package langchaingo
