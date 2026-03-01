package main

import (
	"fmt"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs26"
)

func main() {
	topicMemo := hcs26.BuildTopicMemo(true, 3600, hcs26.TopicTypeDiscovery)
	parsedTopicMemo, ok := hcs26.ParseTopicMemo(topicMemo)
	if !ok {
		panic("failed to parse topic memo")
	}

	transactionMemo := hcs26.BuildTransactionMemo(hcs26.OperationRegister, hcs26.TopicTypeDiscovery)
	parsedTransactionMemo, ok := hcs26.ParseTransactionMemo(transactionMemo)
	if !ok {
		panic("failed to parse transaction memo")
	}

	fmt.Printf("hcs26 topic memo: %+v\n", parsedTopicMemo)
	fmt.Printf("hcs26 tx memo: %+v\n", parsedTransactionMemo)
}
