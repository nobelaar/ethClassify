package main

import (
	"ethClassify/core"
	"fmt"
)

func main() {
	block, err := core.GetBlock()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, tx := range block.Transactions() {
		txType, err := core.Classify(tx)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Tx Type: %s\n\n----------------\n", txType)
	}
}
