package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)//拿到block数据库里拿到blockchain的引用
	UTXOSet := UTXOSet{bc}//从chainstate数据库里拿到utxo的引用
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from) //get a Wallet by its address

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)//组建了一笔交易，并签好了名

	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}//就两笔交易 发起的交易 和奖励给挖矿人的交易 //我觉得没必要

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}