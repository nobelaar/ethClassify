package core

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetBlock() (*types.Block, error) {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/")
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Block Number: ", block.Number())
	fmt.Println("Block Hash: ", block.Hash())
	return block, nil
}
