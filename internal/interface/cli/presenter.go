package cli

import (
	"fmt"
	"math/big"

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
		if tx.Swap != nil {
			fmt.Printf("Swap: dex=%s pair=%s sender=%s recipient=%s a0(in/out)=%s/%s a1(in/out)=%s/%s\n",
				tx.Swap.Dex, tx.Swap.Pair, tx.Swap.Sender, tx.Swap.Recipient,
				formatBigInt(tx.Swap.Amount0In), formatBigInt(tx.Swap.Amount0Out),
				formatBigInt(tx.Swap.Amount1In), formatBigInt(tx.Swap.Amount1Out),
			)
		}
		if tx.Selector != "" {
			fmt.Printf("Function Selector: %s\n", tx.Selector)
		}
		if tx.Details != "" {
			fmt.Printf("Details: %s\n", tx.Details)
		}
		fmt.Println("----------------")
	}
}

func formatBigInt(v *big.Int) string {
	if v == nil {
		return "0"
	}
	return v.String()
}
