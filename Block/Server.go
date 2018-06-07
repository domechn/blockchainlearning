package Block

import (
	"io"
	"fmt"
	"net"
	"log"
	"encoding/gob"
	"bytes"
	"io/ioutil"
	"encoding/hex"
)

type Version struct {
	Version    int
	BestHeight int
	AddFrom    string
}



type GetData struct {
	AddFrom string
	Type    string
	ID      []byte
}

type block struct{
	AddFrom string
	Block []byte
}

type tx struct{
	AddFrom string
	Transaction []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type getblocks struct {
	AddrFrom string
}

type addr struct {
	AddrList []string
}


const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

//通过节点ID和主地址启动服务
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockchain(nodeID)
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	// time.Sleep(time.Second*5)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

//发送版本
func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(Version{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

//读取连接内容
func handleConnection(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
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

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	encoded := gob.NewEncoder(&buff)
	err := encoded.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
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

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytesToCommand(data []byte) string {
	var command []byte
	for _, b := range data {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func sendData(addr string, data []byte) {
	conn , err := net.Dial(protocol,addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _,node := range knownNodes{
			if node != addr {
				updatedNodes = append(updatedNodes,node)
			}
		}

		knownNodes = updatedNodes
		return 
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))

	if err != nil {
		log.Panic(err)
	}
}

//处理当前版本
//如果当前区块链的最高高度是比接收到的高度高那么就将当前版本发送给节点
//如果当前区块不比接收到的高，则向接收地址请求区块
func handleVersion(request []byte, bc *BlockChain) {
	
	var buff bytes.Buffer
	var payload Version
	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		
		sendGetBlocks(payload.AddFrom)
	} else if myBestHeight > foreignerBestHeight {
		
		sendVersion(payload.AddFrom, bc)
	}

	if (!nodeIsKnown(payload.AddFrom)) {
		knownNodes = append(knownNodes, payload.AddFrom)
	}
}

//发送请求区块的通知
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func nodeIsKnown(address string) bool {
	for _, node := range knownNodes {
		if node == address {
			return true
		}
	}

	return false
}

func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
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
	payload := gobEncode(GetData{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock(payload.ID)
		if err != nil {
			log.Panic(err)
		}
		sendBlock(payload.AddFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddFrom, &tx)
	}
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress,b.Serialize()}
	payload:= gobEncode(data)
	request := append(commandToBytes("block"),payload...)

	sendData(addr,request)
}

func sendTx(address string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(address, request)
}

func handleBlock(request []byte,bc *BlockChain){
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

	fmt.Println("Received a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n",block.Hash)

	if len(blocksInTransit) >0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddFrom,"block",blockHash)

		blocksInTransit = blocksInTransit[1:]
	}else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}

}

func handleTx(request []byte,bc *BlockChain){
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
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress ==  knownNodes[0] {
		for _,node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom{
				sendInv(node,"tx",[][]byte{tx.ID})
			}
		}
	}else {
	if len(mempool) >= 2 && len(miningAddress) >0 {
	MineTransactions:
		var txs []*Transaction

		for id := range mempool{
			tx := mempool[id]
			if bc.VerifyTransaction(&tx){
				txs = append(txs,&tx)
			}
		}

		if len(txs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")
			return 
		}

		cbTx := NewCoinbaseTX(miningAddress,"")
		txs = append(txs,cbTx)

		newBlock := bc.MineBlock(txs)

		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()

		fmt.Println("New block is minied!")

		for _,tx := range txs{
			txID := hex.EncodeToString(tx.ID)
			delete(mempool,txID)
		}

		for _,node := range knownNodes {
			if node != nodeAddress{
				sendInv(node ,"block",[][]byte{newBlock.Hash})
			}
		}
		
		if len(mempool) > 0 {
			goto MineTransactions
		}
	}
	}
}