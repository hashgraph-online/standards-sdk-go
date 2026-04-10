package main

import (
	"fmt"

	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs15"
)

func main() {
	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		panic(err)
	}

	tx, err := hcs15.BuildBaseAccountCreateTx(hcs15.BaseAccountCreateTxParams{
		PublicKey:          privateKey.PublicKey(),
		InitialBalanceHbar: 2,
		AccountMemo:        "hcs-15-example",
	})
	if err != nil {
		panic(err)
	}

	_ = tx
	fmt.Printf("built hcs-15 base account transaction for alias %s\n", privateKey.PublicKey().ToEvmAddress())
}
