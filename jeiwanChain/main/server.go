package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

//当一个新的节点开始运行时，它会从一个 DNS 种子获取几个节点，给它们发送 version 消息

//接收到 version 消息的节点会响应自己的 version 消息。
// 这是一种握手：如果没有事先互相问候，就不可能有其他交流。
// 不过，这并不是出于礼貌：version 用于找到一个更长的区块链。
// 当一个节点接收到 version 消息，它会检查本节点的区块链是否比 BestHeight 的值更大。如果不是，节点就会请求并下载缺失的块



const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {//inv 来向其他节点展示当前节点有什么块和交易。再次提醒，它没有包含完整的区块链和交易，仅仅是哈希而已
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

//为了接收消息，我们需要一个服务器
func StartServer(nodeID, minerAddress string) {//每个节点都要启动这个
	nodeAddress=fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress//选择挖矿受益人
	ln, err := net.Listen(protocol, nodeAddress)//
	// 如果是中心节点所有其他节点都会连接到这个节点，这个节点会在其他节点之间发送数据


	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc:=NewBlockchain(nodeID)//连接到本地的区块链数据库

	if nodeAddress != knownNodes[0]{//如果当前节点不是中心节点，它必须向中心节点发送 version 消息来查询是否自己的区块链已过时
		sendVersion(knownNodes[0], bc)//前 12 个字节指定了命令名（比如这里的 version），后面的字节会包含 gob 编码的消息结构
	}

	for {
		conn,err :=ln.Accept()//等待连接
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn,bc)//处理请求
	}
}
func sendVersion(addr string, bc *Blockchain) {//addr:ip发出请求的ip地址
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}


func commandToBytes(command string) []byte {//string 转byte
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {//byte转string
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func gobEncode(data interface{}) []byte {//序列化编码
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func handleConnection(conn net.Conn, bc *Blockchain) {//version处理 输入命令 进行调用
	request, err := ioutil.ReadAll(conn)//从连接中读取request内容
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])//request前12位
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}
func handleGetBlocks(request []byte, bc *Blockchain) {//请求本地所有块的hash 这是个简化版本 todo 单个区块的请求
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])//append request[commandLength:] to buffer 12位之后的指令
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)//对dec进行gob解码并把结果放到payload中
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)//返回block的hash 也可以返回交易的hash
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}//nodeAddress 是本地地址
	payload := gobEncode(inventory)//gob编码序列化
	request := append(commandToBytes("inv"), payload...)//前面是命令的byte类型 后面是gob编码序列化的变量，可以反序列化出一个变量

	sendData(address, request)
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer//声明并初始化了
	var payload inv //声明并出示化了

	buff.Write(request[commandLength:])//append buff 自动变长
	dec := gob.NewDecoder(&buff)//读出buff里的内容到dec，因为要修改buff结构体里的其他变量，因此是*方法
	err := dec.Decode(&payload)//反序列化buff里的内容，copy到payload里 需要修改payload的内容，因此是*方法
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items//浅拷贝，里面的元素指向相同的堆内存

		blockHash := payload.Items[0]//copy 用最新区块的hash值去初始化了blockHash
		sendGetData(payload.AddrFrom, "block", blockHash)//回复接收到的最新区块的hash 请求区块的内容

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {//对每个接收到的区块hash
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)//将待请求区块内容的区块hash暂时存着
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}



func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}
//input:发来请求的地址，准备好的数据string的byte格式
//output:给地址回复准备好的数据
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {//如果出错 更新本地已知节点列表 第一个已知的肯定是自己
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))//发回数据
	if err != nil {
		log.Panic(err)
	}
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)
	//TODO：并非无条件信任，我们应该在将每个块加入到区块链之前对它们进行验证。



	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {//继续请求待传输的区块内容
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}//全传输完了，建立utxo数据库
		UTXOSet.Reindex()
		//TODO: 并非运行 UTXOSet.Reindex()， 而是应该使用 UTXOSet.Update(block)，因为如果区块链很大，它将需要很多时间来对整个 UTXO 集重新索引。这时候还需要调整请求块的顺序
	}
}
func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}


		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		fmt.Println("send tx to ",payload.AddrFrom)
		sendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}


}
func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx//加入到内存池

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})//新的交易推送给网络中的其他节点
			}
		}
	} else {//对于矿工节点
		if len(mempool) >= 2 && len(miningAddress) > 0 {//收到交易后，内存池中有两笔或更多的交易，开始挖矿
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)//添加挖矿奖励

			newBlock := bc.MineBlock(txs)//何时停止挖矿呢？？todo 停止挖矿策略
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {//从内存池里清楚已打包的
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {//通知已知节点
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}
func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}
func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}

}
func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}