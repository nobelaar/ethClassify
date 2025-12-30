package usecase

import (
	"context"
	"fmt"

	"ethClassify/internal/domain"
)

type ClassifyBlock struct {
	Reader       domain.BlockReader
	Classifiers  []domain.TxClassifier
	LogResolvers []domain.TxLogResolver
	Labeler      domain.AddressLabeler
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
			classified = true
			result, err = uc.resolveLogs(ctx, tx, result)
			if err != nil {
				return domain.BlockResult{}, err
			}
			results = append(results, result)
			break
		}

		if !classified {
			result := domain.TxResult{
				Tx:       tx,
				Type:     domain.ClassificationUnknown,
				Selector: selectorHex(tx.Data),
				ToLabel:  labeled,
			}
			result, err = uc.resolveLogs(ctx, tx, result)
			if err != nil {
				return domain.BlockResult{}, err
			}
			results = append(results, result)
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

func (uc ClassifyBlock) resolveLogs(ctx context.Context, tx domain.Tx, current domain.TxResult) (domain.TxResult, error) {
	if len(uc.LogResolvers) == 0 {
		return current, nil
	}

	if current.Type != domain.ClassificationContractCall && current.Type != domain.ClassificationUnknown {
		return current, nil
	}

	if len(tx.Logs) == 0 {
		return current, nil
	}

	resolved := current
	for _, resolver := range uc.LogResolvers {
		next, ok, err := resolver.Resolve(ctx, tx, resolved)
		if err != nil {
			return domain.TxResult{}, err
		}
		if !ok {
			continue
		}
		if next.ToLabel == "" {
			next.ToLabel = resolved.ToLabel
		}
		if next.Tx.Hash == "" {
			next.Tx = tx
		}
		resolved = next
		break
	}

	return resolved, nil
}

func selectorHex(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	return fmt.Sprintf("%x", data[:4])
}
