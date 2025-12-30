package classifier

import (
	"context"

	"ethClassify/internal/domain"
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

type ERC20Classifier struct{}

func (ERC20Classifier) Classify(ctx context.Context, tx domain.Tx) (domain.TxResult, bool, error) {
	if len(tx.Data) < 4 {
		return domain.TxResult{}, false, nil
	}

	selector := selectorHex(tx.Data)
	classType, ok := erc20SelectorMap[selector]
	if !ok {
		return domain.TxResult{}, false, nil
	}

	return domain.TxResult{
		Type:     classType,
		Selector: selector,
	}, true, nil
}

var erc20SelectorMap = map[string]domain.ClassificationType{
	"a9059cbb": domain.ClassificationERC20Transfer,
	"095ea7b3": domain.ClassificationERC20Approve,
	"23b872dd": domain.ClassificationERC20TransferFrom,
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
