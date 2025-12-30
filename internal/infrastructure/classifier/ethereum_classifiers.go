package classifier

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"ethClassify/internal/domain"
)

const (
	transferEventTopic       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	approvalEventTopic       = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
	approvalForAllEventTopic = "0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31"
	uniswapV2SwapTopic       = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	uniswapV3SwapTopic       = "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"
)

type DeployClassifier struct{}

func (DeployClassifier) Classify(ctx context.Context, tx domain.Tx) (domain.TxResult, bool, error) {
	if tx.To != nil {
		return domain.TxResult{}, false, nil
	}
	return domain.TxResult{
		Type:     domain.ClassificationDeploy,
		Selector: selectorHex(tx.Data),
	}, true, nil
}

type NativeTransferClassifier struct{}

func (NativeTransferClassifier) Classify(ctx context.Context, tx domain.Tx) (domain.TxResult, bool, error) {
	if tx.To == nil || tx.Value == nil {
		return domain.TxResult{}, false, nil
	}
	if len(tx.Data) > 0 || tx.Value.Sign() <= 0 {
		return domain.TxResult{}, false, nil
	}
	return domain.TxResult{
		Type:     domain.ClassificationTransfer,
		Selector: "",
	}, true, nil
}

type ContractCallClassifier struct{}

func (ContractCallClassifier) Classify(ctx context.Context, tx domain.Tx) (domain.TxResult, bool, error) {
	if tx.To == nil {
		return domain.TxResult{}, false, nil
	}
	if len(tx.Data) < 4 {
		return domain.TxResult{}, false, nil
	}

	selector := selectorHex(tx.Data)
	return domain.TxResult{
		Type:     domain.ClassificationContractCall,
		Selector: selector,
	}, true, nil
}

var erc20SelectorMap = map[string]domain.ClassificationType{
	"a9059cbb": domain.ClassificationERC20Transfer,
	"095ea7b3": domain.ClassificationERC20Approve,
	"23b872dd": domain.ClassificationERC20TransferFrom,
}

type ERC20LogResolver struct{}

func (ERC20LogResolver) Resolve(ctx context.Context, tx domain.Tx, current domain.TxResult) (domain.TxResult, bool, error) {
	if current.Type != domain.ClassificationContractCall && current.Type != domain.ClassificationUnknown {
		return current, false, nil
	}
	for _, log := range tx.Logs {
		if len(log.Topics) == 0 {
			continue
		}
		if log.Topics[0] == transferEventTopic && len(log.Topics) == 3 {
			updated := current
			updated.Type = erc20TypeFromSelector(current.Selector, domain.ClassificationERC20Transfer)
			return updated, true, nil
		}
		if log.Topics[0] == approvalEventTopic && len(log.Topics) == 3 {
			updated := current
			updated.Type = erc20TypeFromSelector(current.Selector, domain.ClassificationERC20Approve)
			return updated, true, nil
		}
	}
	return current, false, nil
}

type ERC721LogResolver struct{}

func (ERC721LogResolver) Resolve(ctx context.Context, tx domain.Tx, current domain.TxResult) (domain.TxResult, bool, error) {
	if current.Type != domain.ClassificationContractCall && current.Type != domain.ClassificationUnknown {
		return current, false, nil
	}
	for _, log := range tx.Logs {
		if len(log.Topics) == 0 {
			continue
		}
		switch log.Topics[0] {
		case transferEventTopic:
			if len(log.Topics) == 4 {
				updated := current
				updated.Type = domain.ClassificationERC721Transfer
				return updated, true, nil
			}
		case approvalEventTopic:
			if len(log.Topics) == 4 {
				updated := current
				updated.Type = domain.ClassificationERC721Approval
				return updated, true, nil
			}
		case approvalForAllEventTopic:
			if len(log.Topics) == 3 {
				updated := current
				updated.Type = domain.ClassificationERC721ApprovalForAll
				return updated, true, nil
			}
		}
	}
	return current, false, nil
}

type DexSwapLogResolver struct{}

func (DexSwapLogResolver) Resolve(ctx context.Context, tx domain.Tx, current domain.TxResult) (domain.TxResult, bool, error) {
	if current.Type != domain.ClassificationContractCall && current.Type != domain.ClassificationUnknown {
		return current, false, nil
	}
	for _, log := range tx.Logs {
		if len(log.Topics) == 0 {
			continue
		}
		switch log.Topics[0] {
		case uniswapV2SwapTopic:
			swap, ok := parseUniswapV2Swap(log)
			if !ok {
				continue
			}
			updated := current
			updated.Type = domain.ClassificationDexSwap
			updated.Swap = swap
			updated.Details = formatSwapDetails(*swap)
			return updated, true, nil
		case uniswapV3SwapTopic:
			swap, ok := parseUniswapV3Swap(log)
			if !ok {
				continue
			}
			updated := current
			updated.Type = domain.ClassificationDexSwap
			updated.Swap = swap
			updated.Details = formatSwapDetails(*swap)
			return updated, true, nil
		}
	}
	return current, false, nil
}

func erc20TypeFromSelector(selector string, fallback domain.ClassificationType) domain.ClassificationType {
	if classType, ok := erc20SelectorMap[selector]; ok {
		return classType
	}
	return fallback
}

func parseUniswapV2Swap(log domain.Log) (*domain.SwapInfo, bool) {
	if len(log.Data) < 128 {
		return nil, false
	}
	amount0In := new(big.Int).SetBytes(log.Data[0:32])
	amount1In := new(big.Int).SetBytes(log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(log.Data[96:128])

	sender := ""
	recipient := ""
	if len(log.Topics) > 1 {
		sender = topicToAddress(log.Topics[1])
	}
	if len(log.Topics) > 2 {
		recipient = topicToAddress(log.Topics[2])
	}

	return &domain.SwapInfo{
		Dex:        "uniswap-v2",
		Pair:       strings.ToLower(log.Address),
		Sender:     sender,
		Recipient:  recipient,
		Amount0In:  amount0In,
		Amount1In:  amount1In,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
	}, true
}

func parseUniswapV3Swap(log domain.Log) (*domain.SwapInfo, bool) {
	if len(log.Data) < 160 {
		return nil, false
	}
	amount0 := parseSigned256(log.Data[0:32])
	amount1 := parseSigned256(log.Data[32:64])

	amount0In := big.NewInt(0)
	amount1In := big.NewInt(0)
	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(0)

	if amount0.Sign() > 0 {
		amount0In = amount0
	} else if amount0.Sign() < 0 {
		amount0Out = new(big.Int).Neg(amount0)
	}

	if amount1.Sign() > 0 {
		amount1In = amount1
	} else if amount1.Sign() < 0 {
		amount1Out = new(big.Int).Neg(amount1)
	}

	sender := ""
	recipient := ""
	if len(log.Topics) > 1 {
		sender = topicToAddress(log.Topics[1])
	}
	if len(log.Topics) > 2 {
		recipient = topicToAddress(log.Topics[2])
	}

	return &domain.SwapInfo{
		Dex:        "uniswap-v3",
		Pair:       strings.ToLower(log.Address),
		Sender:     sender,
		Recipient:  recipient,
		Amount0In:  amount0In,
		Amount1In:  amount1In,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
	}, true
}

func topicToAddress(topic string) string {
	if len(topic) < 42 {
		return ""
	}
	return strings.ToLower("0x" + topic[len(topic)-40:])
}

func parseSigned256(data []byte) *big.Int {
	if len(data) == 0 {
		return big.NewInt(0)
	}
	v := new(big.Int).SetBytes(data)
	if data[0]&0x80 != 0 {
		max := new(big.Int).Lsh(big.NewInt(1), 256)
		v.Sub(v, max)
	}
	return v
}

func formatSwapDetails(swap domain.SwapInfo) string {
	return fmt.Sprintf("%s swap pair=%s sender=%s a0(in/out)=%s/%s a1(in/out)=%s/%s",
		swap.Dex, swap.Pair, swap.Sender,
		swap.Amount0In.String(), swap.Amount0Out.String(),
		swap.Amount1In.String(), swap.Amount1Out.String(),
	)
}

func selectorHex(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	return string([]byte{
		hexNibble(data[0] >> 4), hexNibble(data[0]),
		hexNibble(data[1] >> 4), hexNibble(data[1]),
		hexNibble(data[2] >> 4), hexNibble(data[2]),
		hexNibble(data[3] >> 4), hexNibble(data[3]),
	})
}

func hexNibble(b byte) byte {
	v := b & 0x0f
	if v < 10 {
		return '0' + v
	}
	return 'a' + (v - 10)
}
