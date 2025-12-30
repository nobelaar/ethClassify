package classifier

import (
	"context"

	"ethClassify/internal/domain"
)

const (
	transferEventTopic       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	approvalEventTopic       = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
	approvalForAllEventTopic = "0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31"
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

func erc20TypeFromSelector(selector string, fallback domain.ClassificationType) domain.ClassificationType {
	if classType, ok := erc20SelectorMap[selector]; ok {
		return classType
	}
	return fallback
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
