package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs11"
)

func main() {
	client, err := hcs11.NewClient(hcs11.ClientConfig{
		Network: "testnet",
	})
	if err != nil {
		panic(err)
	}

	profile, err := client.CreateAIAgentProfile(
		"Support Agent",
		hcs11.AIAgentTypeAutonomous,
		[]hcs11.AIAgentCapability{
			hcs11.AIAgentCapabilityTextGeneration,
			hcs11.AIAgentCapabilityKnowledgeRetrieval,
		},
		"gpt-5",
		map[string]any{
			"alias": "support-agent",
		},
	)
	if err != nil {
		panic(err)
	}

	encoded, err := client.ProfileToJSONString(profile)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded)
}
