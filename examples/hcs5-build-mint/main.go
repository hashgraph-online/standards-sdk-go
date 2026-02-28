package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs5"
)

func main() {
	tokenID := "0.0.123456"
	metadataTopicID := "0.0.654321"

	tx, err := hcs5.BuildMintWithHRLTx(tokenID, metadataTopicID, "")
	if err != nil {
		panic(err)
	}

	_ = tx
	fmt.Printf("built hcs-5 mint transaction for token %s with metadata %s\n", tokenID, hcs5.BuildHCS1HRL(metadataTopicID))
}
