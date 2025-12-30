package main

import (
	"context"
	"flag"
	"log"

	"ethClassify/internal/domain"
	"ethClassify/internal/infrastructure/classifier"
	"ethClassify/internal/infrastructure/ethereum"
	"ethClassify/internal/infrastructure/labeler"
	"ethClassify/internal/interface/cli"
	"ethClassify/internal/usecase"
)

func main() {
	url := flag.String("url", "", "rpc url raw link")
	flag.Parse()
	if *url == "" {
		log.Fatal("rpc-url is required")
	}

	reader, err := ethereum.NewBlockReader(*url)
	if err != nil {
		log.Fatalf("failed to create block reader: %v", err)
	}

	addrLabeler := labeler.NewStaticLabeler(map[string]string{
		"0xdac17f958d2ee523a2206206994597c13d831ec7": "USDT",
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "USDC",
		"0x6b175474e89094c44da98b954eedeac495271d0f": "DAI",
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": "WETH",
	})

	classifiers := []domain.TxClassifier{
		classifier.DeployClassifier{},
		classifier.NativeTransferClassifier{},
		classifier.ContractCallClassifier{},
	}

	resolvers := []domain.TxLogResolver{
		classifier.ERC721LogResolver{},
		classifier.ERC20LogResolver{},
	}

	uc := usecase.ClassifyBlock{
		Reader:       reader,
		Classifiers:  classifiers,
		LogResolvers: resolvers,
		Labeler:      addrLabeler,
	}

	ctx := context.Background()
	result, err := uc.Execute(ctx)
	if err != nil {
		log.Fatalf("failed to classify block: %v", err)
	}

	cli.PrintBlockResult(result)
}
