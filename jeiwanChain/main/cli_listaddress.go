package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses(nodeID string) {//读出本地拥有的钱包地址
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}