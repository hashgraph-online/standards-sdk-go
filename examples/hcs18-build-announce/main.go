package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs18"
)

func main() {
	message := hcs18.BuildAnnounceMessage(hcs18.AnnounceData{
		Account: "0.0.123",
		Petal: hcs18.PetalDescriptor{
			Name:     "agent",
			Priority: 1,
		},
		Capabilities: hcs18.CapabilityDetails{
			Protocols: []string{"hcs-10"},
		},
	})
	transaction, err := hcs18.BuildSubmitDiscoveryMessageTx("0.0.123", message, "")
	if err != nil {
		panic(err)
	}
	fmt.Printf("built hcs18 announce tx: %T\n", transaction)
}
