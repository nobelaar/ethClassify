package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"ethClassify/internal/domain"
	"ethClassify/internal/infrastructure/classifier"
	"ethClassify/internal/infrastructure/ethereum"
	"ethClassify/internal/infrastructure/labeler"
	"ethClassify/internal/interface/cli"
	"ethClassify/internal/usecase"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "ethClassify - clasificador de transacciones de Ethereum")
		fmt.Fprintf(flag.CommandLine.Output(), "Uso: %s -url <rpc-url> [opciones]\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Clasifica el ultimo bloque de Ethereum mainnet usando el endpoint RPC indicado.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOpciones:")
		fmt.Fprintln(flag.CommandLine.Output(), "\t-url <rpc-url>\tRPC URL")
		fmt.Fprintln(flag.CommandLine.Output(), "\t-with-logs\tUsa logs para clasificar transacciones ERC (hace más llamadas RPC!!)")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nEjemplo:\n  %s -url https://mainnet.infura.io/v3/<project-id> -with-logs", os.Args[0])
	}

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	url := flag.String("url", "", "rpc url raw link")
	withLogs := flag.Bool("with-logs", false, "use transaction receipts/logs for ERC-type classification (extra RPC calls)")
	flag.Parse()
	if *url == "" {
		fmt.Fprintln(flag.CommandLine.Output(), "error: -url is required")
		flag.Usage()
		os.Exit(2)
	}

	reader, err := ethereum.NewBlockReader(*url, *withLogs)
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

	var resolvers []domain.TxLogResolver
	if *withLogs {
		resolvers = []domain.TxLogResolver{
			classifier.ERC721LogResolver{},
			classifier.ERC20LogResolver{},
		}
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
