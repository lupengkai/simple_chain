package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp,10))//转换格式
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})//数据相连
	hash :=sha256.Sum256(headers)//hash函数

	b.Hash = hash[:]//hash是[32]byte类型 b.Hash是[]byte类型
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(),[]byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}