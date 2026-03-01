package langchaingo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs14"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

// UAIDResolverTool is a langchaingo compatible Tool that allows an agent
// to resolve a HOL Universal Agent ID (UAID) to its profile.
type UAIDResolverTool struct {
	client    *hcs14.Client
	Callbacks callbacks.Handler
}

// Ensure UAIDResolverTool implements tools.Tool
var _ tools.Tool = &UAIDResolverTool{}

// NewUAIDResolverTool creates a new langchaingo tool for resolving UAIDs.
// If client is nil, it initializes a default HCS-14 client.
func NewUAIDResolverTool(client *hcs14.Client) *UAIDResolverTool {
	if client == nil {
		client = hcs14.NewClient(hcs14.ClientOptions{})
	}
	return &UAIDResolverTool{
		client: client,
	}
}

// Name returns the name of the tool.
func (t *UAIDResolverTool) Name() string {
	return "Hashgraph_Online_UAID_Resolver"
}

// Description returns a description of the tool to help the language model
// decide when to use it.
func (t *UAIDResolverTool) Description() string {
	return `Resolves a HOL Universal Agent ID (UAID) string (e.g., uaid:aid:...) into its detailed JSON profile. 
Use this tool when you need to look up information, authentication details, or endpoints for an AI agent given its UAID.`
}

// Call executes the tool by resolving the provided UAID string.
// Input should be a valid UAID string.
func (t *UAIDResolverTool) Call(ctx context.Context, input string) (string, error) {
	if t.Callbacks != nil {
		t.Callbacks.HandleToolStart(ctx, input)
	}

	result, err := t.client.Resolve(ctx, input)
	if err != nil {
		if t.Callbacks != nil {
			t.Callbacks.HandleToolError(ctx, err)
		}
		return fmt.Sprintf("Failed to resolve UAID: %v", err), nil
	}

	// Format result as pretty JSON to feed back to the LLM agent
	jsonData, err := json.MarshalIndent(result.Metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode profile to JSON: %w", err)
	}

	output := string(jsonData)

	if t.Callbacks != nil {
		t.Callbacks.HandleToolEnd(ctx, output)
	}

	return output, nil
}
