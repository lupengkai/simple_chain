package main

import "bytes"



type TXOutput struct {
	Value      int
	PubKeyHash []byte //等于是转到的地址
}

func (out *TXOutput) Lock(address []byte) {//由地址得到pubKeyHash
	pubKeyHash := Base58Decode(address)//得到version+Publikey+checsum
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {//检查是否提供的公钥哈希是否和tx里的公钥 hash一样
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}






func (output TXOutput) CanBeUnlockedWith(s string) bool {

}
