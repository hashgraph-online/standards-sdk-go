package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs16"
)

func main() {
	floraAccountID := "0.0.500001"

	tx, err := hcs16.BuildCreateFloraTopicTx(hcs16.CreateFloraTopicOptions{
		FloraAccountID: floraAccountID,
		TopicType:      hcs16.FloraTopicTypeCommunication,
	})
	if err != nil {
		panic(err)
	}

	_ = tx
	fmt.Printf("built hcs-16 flora topic create transaction for flora account %s\n", floraAccountID)
}
