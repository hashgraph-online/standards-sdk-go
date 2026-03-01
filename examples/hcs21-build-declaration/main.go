package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs21"
)

func main() {
	declaration := hcs21.AdapterDeclaration{
		P:         "hcs-21",
		Op:        hcs21.OperationRegister,
		AdapterID: "demo-adapter",
		Entity:    "service",
		Package: hcs21.AdapterPackage{
			Registry:  "npm",
			Name:      "demo-adapter",
			Version:   "1.0.0",
			Integrity: "sha384-demo",
		},
		Manifest: "hcs://1/0.0.123",
		Config: map[string]any{
			"type": "state",
		},
	}

	transaction, err := hcs21.BuildDeclarationMessageTx("0.0.123", declaration, "")
	if err != nil {
		panic(err)
	}
	fmt.Printf("built hcs21 declaration tx: %T\n", transaction)
}
