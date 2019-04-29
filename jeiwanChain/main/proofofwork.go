package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)


var (
	maxNonce = math.MaxInt64
)

const targetBits = 16 // 计算难度

type ProofOfWork struct {
	block *Block
	target *big.Int //64位
}

func NewProofOfWork(b *Block) *ProofOfWork { //不是实例方法
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits)) //target 左移 256-targetbits位 这个可以用来比较难度

	pow := &ProofOfWork{b, target}

	return pow

}

/*func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n,16))//转化为string类型
}*/
func (pow *ProofOfWork) prepareData(nonce int) []byte {//私有方法
	data :=bytes.Join(//前一个区块的hash，区块包含的数据，时间戳，难度，nonce值
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
		)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {//公开方法
	var hashInt big.Int
	var hash [32]byte
	nonce := 0 //nonce初始化

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)//16进制表示
		if math.Remainder(float64(nonce), 100000) == 0 {
			fmt.Printf("\r%x", hash)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
 		} else {
 			nonce++//nonce一个一个试
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:] //可能找到了满足条件的， 也可能是最大值
}


func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data :=pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid

}

/*func main() {
	data1 := []byte("I like donuts")
	data2 := []byte("I like donutsca07ca")
	targetBits :=24
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	fmt.Printf("%x\n", sha256.Sum256(data1))
	fmt.Printf("%64x\n", target)
	fmt.Printf("%x\n", sha256.Sum256((data2)))
}
*/