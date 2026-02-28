package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs14"
)

func main() {
	uaid := "uaid:aid:ans-godaddy-ote;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1"

	parsed, err := hcs14.ParseUAID(uaid)
	if err != nil {
		panic(err)
	}

	canonical := hcs14.BuildCanonicalUAID(parsed.Target, parsed.ID, parsed.Params)

	fmt.Printf("target=%s\n", parsed.Target)
	fmt.Printf("id=%s\n", parsed.ID)
	fmt.Printf("canonical=%s\n", canonical)
}
