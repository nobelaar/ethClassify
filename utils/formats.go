package utils

import "math/big"

func WeiToEtherString(wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	eth := new(big.Float).Quo(f, big.NewFloat(1e18))
	return eth.Text('f', 18)
}
