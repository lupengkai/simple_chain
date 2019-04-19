package main

import "bytes"

type TXInput struct {
	Txid []byte
	Vout int
	Signature []byte
	PubKey []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool { //检查输入使用了指定密钥来解锁一个输出
	lockingHash := HashPubKey(in.PubKey) // 两次hash后得到的publickey hash
	return bytes.Compare(lockingHash, pubKeyHash) == 0

}
