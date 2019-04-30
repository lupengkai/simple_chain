package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address, nodeID string) {//由cli创建区块链，
	if !ValidateAddress(address) {//验证是否为合法地址
		log.Panic("ERROR: Address is not valid")
	}
	//todo  硬编码创世区块
	bc := CreateBlockchain(address, nodeID)//根据地址生成创世区块，同时创建block数据库 并持久化创世区块到block数据库
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}//创建UTXO对象，管理维护区块链的账户余额
	UTXOSet.Reindex()//创建chainstate bucket，并根据创世区块更新chainstate

	fmt.Println("Done!")
}
