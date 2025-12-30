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
	Logs  []Log
}

type ClassificationType string

const (
	ClassificationDeploy               ClassificationType = "DEPLOY"
	ClassificationTransfer             ClassificationType = "TRANSFER"
	ClassificationContractCall         ClassificationType = "CONTRACT_CALL"
	ClassificationDexSwap              ClassificationType = "DEX_SWAP"
	ClassificationSandwichSuspect      ClassificationType = "SANDWICH_SUSPECT"
	ClassificationERC20Transfer        ClassificationType = "ERC20_TRANSFER"
	ClassificationERC20Approve         ClassificationType = "ERC20_APPROVE"
	ClassificationERC20TransferFrom    ClassificationType = "ERC20_TRANSFER_FROM"
	ClassificationERC721Transfer       ClassificationType = "ERC721_TRANSFER"
	ClassificationERC721Approval       ClassificationType = "ERC721_APPROVAL"
	ClassificationERC721ApprovalForAll ClassificationType = "ERC721_APPROVAL_FOR_ALL"
	ClassificationUnknown              ClassificationType = "UNKNOWN"
)

type TxResult struct {
	Tx       Tx
	Type     ClassificationType
	Selector string
	ToLabel  string
	Swap     *SwapInfo
	Details  string
}

type BlockResult struct {
	Block   Block
	Results []TxResult
}

type SwapInfo struct {
	Dex        string
	Pair       string
	Sender     string
	Recipient  string
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
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

type TxLogResolver interface {
	Resolve(ctx context.Context, tx Tx, current TxResult) (TxResult, bool, error)
}

type Log struct {
	Address string
	Topics  []string
	Data    []byte
}
