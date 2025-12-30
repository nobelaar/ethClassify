package usecase

import (
	"context"
	"fmt"
	"math/big"

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

	results = markSandwiches(results)

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

func markSandwiches(results []domain.TxResult) []domain.TxResult {
	if len(results) < 3 {
		return results
	}

	for i := 1; i < len(results)-1; i++ {
		pre := results[i-1]
		victim := results[i]
		post := results[i+1]

		preFlow, ok := swapFlow(pre)
		if !ok {
			continue
		}
		victimFlow, ok := swapFlow(victim)
		if !ok {
			continue
		}
		postFlow, ok := swapFlow(post)
		if !ok {
			continue
		}

		if preFlow.pair != victimFlow.pair || preFlow.pair != postFlow.pair {
			continue
		}
		if preFlow.sender == "" || preFlow.sender != postFlow.sender {
			continue
		}
		if victimFlow.sender == preFlow.sender {
			continue
		}
		if preFlow.direction != victimFlow.direction || preFlow.direction == postFlow.direction {
			continue
		}
		if !similarAmount(preFlow.outAmount, postFlow.inAmount, 30) {
			continue
		}

		results[i].Type = domain.ClassificationSandwichSuspect
		results[i].Details = fmt.Sprintf("Possible sandwich: frontrun %s / backrun %s attacker %s", pre.Tx.Hash, post.Tx.Hash, preFlow.sender)
	}

	return results
}

type flow struct {
	pair      string
	sender    string
	direction int // 0 -> token1, 1 -> token0
	inAmount  *big.Int
	outAmount *big.Int
}

func swapFlow(res domain.TxResult) (flow, bool) {
	if res.Swap == nil {
		return flow{}, false
	}

	switch {
	case res.Swap.Amount0In != nil && res.Swap.Amount0In.Sign() > 0 && res.Swap.Amount1Out != nil && res.Swap.Amount1Out.Sign() > 0:
		return flow{
			pair:      res.Swap.Pair,
			sender:    res.Swap.Sender,
			direction: 0,
			inAmount:  res.Swap.Amount0In,
			outAmount: res.Swap.Amount1Out,
		}, true
	case res.Swap.Amount1In != nil && res.Swap.Amount1In.Sign() > 0 && res.Swap.Amount0Out != nil && res.Swap.Amount0Out.Sign() > 0:
		return flow{
			pair:      res.Swap.Pair,
			sender:    res.Swap.Sender,
			direction: 1,
			inAmount:  res.Swap.Amount1In,
			outAmount: res.Swap.Amount0Out,
		}, true
	default:
		return flow{}, false
	}
}

func similarAmount(a, b *big.Int, tolerancePercent int64) bool {
	if a == nil || b == nil || a.Sign() <= 0 || b.Sign() <= 0 {
		return false
	}
	diff := new(big.Int).Sub(a, b)
	if diff.Sign() < 0 {
		diff.Neg(diff)
	}
	threshold := new(big.Int).Mul(a, big.NewInt(tolerancePercent))
	threshold.Div(threshold, big.NewInt(100))
	return diff.Cmp(threshold) <= 0
}
