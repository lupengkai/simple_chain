package main
//等于Account 保存chain state 的对象
import(
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucket = "chainstate"
//chainstate 存储未花费交易输出的集合  数据库表示的未花费交易输出的块哈希     交易哈希-》块哈希-》块

type UTXOSet struct { //UTXO(chainstate) 跟 blocks一起 但存储在不同的 bucket 中
	Blockchain *Blockchain
}

func (u UTXOSet) Reindex() {//初始化的时候使用一次 遍历block数据库也是初始化的时候使用一次
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket) //拿到数据源

	err := db.Update(func(tx *bolt.Tx) error {//在数据库中执行事务：新建一个bucket
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}
		return nil//事务里的return
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO() //从区块链里找到所有的未花费交易  TransactionID -> TransactionOutputs 的 map
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {//将未花费交易 以 id(即hash): outputs （value, publichash） 存到数据库里
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {//遍历chainstate数据库去找
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {//遍历chainstate数据库去找
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain

//我们的数据（交易）现在已经被分开存储：实际交易被存储在区块链中，未花费输出被存储在 UTXO 集中。
// 我们就需要一个良好的同步机制，因为我们想要 UTXO 集时刻处于最新状态，并且存储最新交易的输出。
// 但是我们不想每生成一个新块，就重新生成索引，因为这正是我们要极力避免的频繁区块链扫描。
// 因此，我们需要一个机制来更新 UTXO 集
func (u UTXOSet) Update(block *Block) {//最新加进来的区块
	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {//对其中的每一笔交易
			if tx.IsCoinbase() == false {//非coinbase
				for _, vin := range tx.Vin {//处理每个输入
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)//找到utxo里含这次被花掉的输入的事务
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {//遍历未花费完的事务里的输入进行遍历
						if outIdx != vin.Vout {//vin.Vout 该笔输入的来源交易中的索引编号
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)//没有被花掉的输出保存下来
						}
					}

					if len(updatedOuts.Outputs) == 0 { //如果一个transacton下的所有output都被用了 就删除
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize()) //处理完用掉的输出 把记录保存回数据库
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			//接着处理新产生的交易输出
			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {//对每个交易下的每个新产生的输出
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())//存入数据库
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

