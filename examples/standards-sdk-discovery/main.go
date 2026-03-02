package main

import (
	"fmt"

	standardssdk "github.com/hashgraph-online/standards-sdk-go/standards-sdk"
)

func main() {
	fmt.Printf("package=%s\n", standardssdk.PackageName)
	fmt.Printf("import=%s\n", standardssdk.Identity())
}
