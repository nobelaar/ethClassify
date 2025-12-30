package cli

import (
	"fmt"

	"ethClassify/internal/domain"
	"ethClassify/utils"
)

func PrintBlockResult(result domain.BlockResult) {
	fmt.Printf("Block Number: %s\n", result.Block.Number)
	fmt.Printf("Block Hash:   %s\n", result.Block.Hash)

	for _, tx := range result.Results {
		fmt.Println()
		fmt.Printf("Tx Hash: %s\n", tx.Tx.Hash)

		to := "CONTRACT_CREATION"
		if tx.Tx.To != nil {
			to = *tx.Tx.To
			if tx.ToLabel != "" {
				to = fmt.Sprintf("%s (%s)", to, tx.ToLabel)
			}
		}
		fmt.Printf("Tx To: %s\n", to)

		value := ""
		if tx.Tx.Value != nil {
			value = fmt.Sprintf("%s wei (%s ETH)", tx.Tx.Value, utils.WeiToEtherString(tx.Tx.Value))
		}
		fmt.Printf("Tx Value: %s\n", value)
		fmt.Printf("Tx Data: %x\n", tx.Tx.Data)
		fmt.Printf("Classification: %s\n", tx.Type)
		if tx.Selector != "" {
			fmt.Printf("Function Selector: %s\n", tx.Selector)
		}
		if tx.Details != "" {
			fmt.Printf("Details: %s\n", tx.Details)
		}
		fmt.Println("----------------")
	}
}
