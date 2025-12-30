package core

import (
	"fmt"

	"encoding/hex"

	"ethClassify/utils"

	"github.com/ethereum/go-ethereum/core/types"
)

func Classify(tx *types.Transaction) (string, error) {
	to := tx.To()
	fmt.Println()
	fmt.Println("Tx Hash: ", tx.Hash())
	if to == nil {
		fmt.Println("Tx To: CONTRACT_CREATION")
	} else {
		fmt.Println("Tx To: ", ClassifyAddr(to))
	}
	fmt.Printf("Tx Value: %s (%s ETH)\n", tx.Value(), utils.WeiToEtherString(tx.Value()))
	fmt.Println("Tx Data: ", hex.EncodeToString(tx.Data()))
	// contrato nuevo
	if to == nil {
		return "DEPLOY", nil
	}
	// transferencia
	if len(tx.Data()) == 0 && tx.Value().Sign() > 0 {
		return "TRANSFER", nil
	}

	if len(tx.Data()) >= 4 {
		selector := tx.Data()[:4]
		switch fmt.Sprintf("%x", selector) {
		case "a9059cbb":
			return "ERC_TRANSFER", nil
		case "095ea7b3":
			return "ERC_APPROVE", nil
		case "23b872dd":
			return "ERC_TRANSFER_FROM", nil
		default:
			return "UNKNOWN", nil
		}
	}
	return "UNKNOWN", nil
}
