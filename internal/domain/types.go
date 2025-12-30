package domain

import (
	"context"
	"math/big"
)

type Block struct {
	Number       *big.Int
	Hash         string
	Transactions []Tx
}

type Tx struct {
	Hash  string
	To    *string
	Value *big.Int
	Data  []byte
}

type ClassificationType string

const (
	ClassificationDeploy            ClassificationType = "DEPLOY"
	ClassificationTransfer          ClassificationType = "TRANSFER"
	ClassificationERC20Transfer     ClassificationType = "ERC20_TRANSFER"
	ClassificationERC20Approve      ClassificationType = "ERC20_APPROVE"
	ClassificationERC20TransferFrom ClassificationType = "ERC20_TRANSFER_FROM"
	ClassificationUnknown           ClassificationType = "UNKNOWN"
)

type TxResult struct {
	Tx       Tx
	Type     ClassificationType
	Selector string
	ToLabel  string
	Details  string
}

type BlockResult struct {
	Block   Block
	Results []TxResult
}

type BlockReader interface {
	LatestBlock(ctx context.Context) (Block, error)
}

type AddressLabeler interface {
	Label(addr string) string
}

type TxClassifier interface {
	Classify(ctx context.Context, tx Tx) (TxResult, bool, error)
}
