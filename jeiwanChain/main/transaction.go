package main

import (
	"encoding/hex"

	"log"
)

type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

func (tx *Transaction) SetID() { //指针就收对象可以修改值 非指针不能修改

}

func NewUTXOTransaction(from string, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount) //寻找没有花费的输出

	if acc < amount { //acc 是 之前utxo之和
		log.Panic("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs{//遍历utxo
		txID, err :=hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {//组装交易
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, TXOutput{amount, to})//付钱
	if acc >amount {
		outputs = append(outputs, TXOutput{acc-amount,from}) //找零
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}

