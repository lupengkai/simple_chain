package main

import (
	"bytes"
	"encoding/gob"
	"log"
)



type TXOutput struct {
	Value      int
	PubKeyHash []byte //等于是转到的地址
}

type TXOutputs struct {
	Outputs []TXOutput
}


func (out *TXOutput) Lock(address []byte) {//由地址得到pubKeyHash
	pubKeyHash := Base58Decode(address)//得到version+Publikey+checsum
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {//检查是否提供的公钥哈希是否和tx里的公钥 hash一样
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}




func (output TXOutput) CanBeUnlockedWith(s string) bool {

}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
