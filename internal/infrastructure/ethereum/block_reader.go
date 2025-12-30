package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"ethClassify/internal/domain"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockReader struct {
	client *ethclient.Client
}

func NewBlockReader(rpcURL string) (*BlockReader, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect rpc: %w", err)
	}
	return &BlockReader{client: client}, nil
}

func (r *BlockReader) LatestBlock(ctx context.Context) (domain.Block, error) {
	if r == nil || r.client == nil {
		return domain.Block{}, fmt.Errorf("rpc client is not initialized")
	}
	block, err := r.client.BlockByNumber(ctx, nil)
	if err != nil {
		return domain.Block{}, fmt.Errorf("fetch latest block: %w", err)
	}

	return convertBlock(block), nil
}

func convertBlock(block *types.Block) domain.Block {
	txns := block.Transactions()
	out := make([]domain.Tx, 0, len(txns))
	for _, tx := range txns {
		out = append(out, convertTx(tx))
	}

	return domain.Block{
		Number:       new(big.Int).Set(block.Number()),
		Hash:         block.Hash().Hex(),
		Transactions: out,
	}
}

func convertTx(tx *types.Transaction) domain.Tx {
	var toStr *string
	if tx.To() != nil {
		addr := tx.To().Hex()
		toStr = &addr
	}

	return domain.Tx{
		Hash:  tx.Hash().Hex(),
		To:    toStr,
		Value: new(big.Int).Set(tx.Value()),
		Data:  append([]byte(nil), tx.Data()...),
	}
}
