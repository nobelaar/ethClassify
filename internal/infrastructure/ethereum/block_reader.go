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

	return convertBlock(ctx, r.client, block)
}

func convertBlock(ctx context.Context, client *ethclient.Client, block *types.Block) (domain.Block, error) {
	txns := block.Transactions()
	out := make([]domain.Tx, 0, len(txns))
	for _, tx := range txns {
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return domain.Block{}, fmt.Errorf("fetch receipt for tx %s: %w", tx.Hash(), err)
		}
		out = append(out, convertTx(tx, receipt))
	}

	return domain.Block{
		Number:       new(big.Int).Set(block.Number()),
		Hash:         block.Hash().Hex(),
		Transactions: out,
	}, nil
}

func convertTx(tx *types.Transaction, receipt *types.Receipt) domain.Tx {
	var toStr *string
	if tx.To() != nil {
		addr := tx.To().Hex()
		toStr = &addr
	}

	logs := make([]domain.Log, 0, len(receipt.Logs))
	for _, l := range receipt.Logs {
		topics := make([]string, len(l.Topics))
		for i, t := range l.Topics {
			topics[i] = t.Hex()
		}
		logs = append(logs, domain.Log{
			Address: l.Address.Hex(),
			Topics:  topics,
			Data:    append([]byte(nil), l.Data...),
		})
	}

	return domain.Tx{
		Hash:  tx.Hash().Hex(),
		To:    toStr,
		Value: new(big.Int).Set(tx.Value()),
		Data:  append([]byte(nil), tx.Data()...),
		Logs:  logs,
	}
}
