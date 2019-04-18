package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

/*func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp,10))//转换格式
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})//数据相连
	hash :=sha256.Sum256(headers)//hash函数

	b.Hash = hash[:]//hash是[32]byte类型 b.Hash是[]byte类型
}*/

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(),[]byte(data), prevBlockHash, []byte{},0}//打包时生成的时间
	pow := NewProofOfWork(block)//这时候没有nonce
	nonce, hash:=pow.Run()//找一个满足条件的nonce并返回

	block.Hash = hash[:]//填上现在的hash
	block.Nonce = nonce//填上nonce
	return block
}

func (b *Block) Serialize() []byte { //序列化
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err:=encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}


