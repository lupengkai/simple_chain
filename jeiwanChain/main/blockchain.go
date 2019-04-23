package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain_%s.db"
const blocksBucket="blocks"
const genesisCoinbaseData ="The time 18/April/2019 Chancellor on brink of second bailout for banks"



type Blockchain struct {
	tip []byte //指向最后一个区块
	db *bolt.DB //变量名小写表示private
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b :=tx.Bucket([]byte(blocksBucket)) //找到表
		err := b.Put(newBlock.Hash, newBlock.Serialize())//放入
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)//修改l指向最新块的hash
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}



}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b== nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{tip, db}
	return &bc
}


func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}


func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int,map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)//特定账户的unspent txs
	accumulated :=0
	Work:
		for _, tx := range unspentTXs { //每个tx里有几个output
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				if out.CanBeUnlockedWith(address) && accumulated < amount { //找到对应的outputid，且累积起来的未花费金额刚好超过需要转的钱
					accumulated += out.Value
					unspentOutputs[txID]=append(unspentOutputs[txID], outIdx)
					if accumulated >= amount {
						break Work//这样一下子跳出两层循环
				}

				}
			}

		}

		return accumulated, unspentOutputs //返回txid 和 金额

	
	

}

func (bc *Blockchain) FindUnspentTransactions(s string) []Transaction {//找到含特定地址为输出的未花费交易
	
}
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	newBlock := NewBlock(transactions, lastHash)
}


// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {//(*tx).id 也行 这里是省略了
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {//找到所有未花费交易
	UTXO := make(map[string]TXOutputs)//make 分配空间
	spentTXOs := make(map[string][]int)// 一个花掉的 一个没有花掉的
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {//每个交易
			txID := hex.EncodeToString(tx.ID)//由hash之后的byte变成string

		Outputs:
			for outIdx, out := range tx.Vout {//每笔输出
				// Was the output spent?
				if spentTXOs[txID] != nil {//当前交易中有输出被花掉了 才有记录      [[txID-1:outIDx-1,outIDx-2,outIDx-3], [txID-2:outIDx-1,outIDx-2,outIDx-3]...]
					for _, spentOutIdx := range spentTXOs[txID] {//被记录过的花掉的输出
						if spentOutIdx == outIdx {//当前的输出被花掉了，没啥好说的 跳到下笔输出
							continue Outputs//下笔输出
						}
					}
				}
				//当前的输出没有被花掉的话，则
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)//out(value, pubkeyhash)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {//coincase 的输入不需要管
				for _, in := range tx.Vin { //Vin (id, value, pubhash, signature)
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)//把来源的trantion id和 output
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO//transaction的id 以及金额 和 对象
}


/*func main() { // 邮件 main包 run
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.blocks{
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Println("PoW: %s \n", strconv.FormatBool(pow.Validate())) //%t 是bool类型的占位符
	}
}*/
