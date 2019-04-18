package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	currentHash []byte //当前指针
	db *bolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block
	err := i.db.View(func(tx *bolt.Tx) error { //根据hash从数据库里读取制定的区块
		b:= tx.Bucket([]byte(blocksBucket))
		encodeBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodeBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevBlockHash //当前指针往前挪一个

	return block
}

