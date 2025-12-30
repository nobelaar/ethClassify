package usecase

import (
	"context"
	"fmt"

	"ethClassify/internal/domain"
)

type ClassifyBlock struct {
	Reader      domain.BlockReader
	Classifiers []domain.TxClassifier
	Labeler     domain.AddressLabeler
}

func (uc ClassifyBlock) Execute(ctx context.Context) (domain.BlockResult, error) {
	if uc.Reader == nil {
		return domain.BlockResult{}, fmt.Errorf("block reader is required")
	}
	if len(uc.Classifiers) == 0 {
		return domain.BlockResult{}, fmt.Errorf("at least one classifier is required")
	}

	block, err := uc.Reader.LatestBlock(ctx)
	if err != nil {
		return domain.BlockResult{}, err
	}

	results := make([]domain.TxResult, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		labeled := uc.label(tx.To)
		classified := false

		for _, classifier := range uc.Classifiers {
			result, ok, err := classifier.Classify(ctx, tx)
			if err != nil {
				return domain.BlockResult{}, err
			}
			if !ok {
				continue
			}
			result.Tx = tx
			result.ToLabel = labeled
			results = append(results, result)
			classified = true
			break
		}

		if !classified {
			results = append(results, domain.TxResult{
				Tx:       tx,
				Type:     domain.ClassificationUnknown,
				Selector: selectorHex(tx.Data),
				ToLabel:  labeled,
			})
		}
	}

	return domain.BlockResult{
		Block:   block,
		Results: results,
	}, nil
}

func (uc ClassifyBlock) label(addr *string) string {
	if addr == nil || uc.Labeler == nil {
		return ""
	}
	return uc.Labeler.Label(*addr)
}

func selectorHex(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	return fmt.Sprintf("%x", data[:4])
}
