package main

import (
	"ethClassify/core"
	"flag"
	"fmt"
	"os"
)

func main() {
	url := flag.String("url", "", "rpc url raw link")
	flag.Parse()
	if *url == "" {
		fmt.Println("rpc-url is required")
		os.Exit(1)
	}
	block, err := core.GetBlock(*url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, tx := range block.Transactions() {
		txType, err := core.Classify(tx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Tx Type: %s\n\n----------------\n", txType)
	}
}
