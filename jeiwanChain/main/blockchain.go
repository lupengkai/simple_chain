package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
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

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}



// 从block数据库里返回区块链的引用，通过该引用可以管理block
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}



// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address, nodeID string) *Blockchain {//创建并初始化区block块链数据库
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {//检查本地是否有blockchain的数据库文件 有就退出
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)//新建创世区块

	//连接数据库 持久化创世区块
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}


// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)//将这次交易使用的输出来源的交易找到
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


// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true { //现在是如果块里面有一个非法的就放弃整个块里的交易
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {//得到前面一个块的hash和高度
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}
// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
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
// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {//返回本地的区块链所有区块的hash
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)//[4,3,2,1] 最新的在最前面

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}
