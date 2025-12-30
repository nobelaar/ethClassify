package core

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func ClassifyAddr(addr *common.Address) string {
	if addr == nil {
		return "nil"
	}
	addrStr := strings.ToLower(addr.String())
	switch addrStr {
	case "0xdac17f958d2ee523a2206206994597c13d831ec7":
		return fmt.Sprintf("%s (%s)", addrStr, "USDT Contract")
	}
	return fmt.Sprintf("%s", addrStr)
}
