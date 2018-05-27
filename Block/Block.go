package Block

import (
	"time"

	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
)

//定义一个区块链结构
type Block struct{
	Timestamp int64  //时间戳
	Transactions []*Transaction  //携带数据
	PrevHash []byte  //前一块哈希值
	Hash []byte //哈希
	Nonce int //工作量证明
}


//生成一个区块并计算它的哈希值和计算量证明
func NewBlock(transactions []*Transaction,prevBlockHash []byte) (block *Block) {
	//创建一个区块
	block = &Block{time.Now().Unix(),transactions,prevBlockHash,[]byte{},0}
	//创建一个新的POW
	pow := NewproofOfWork(block)
	//生成哈希值和工作量证明
	nonce , hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return 
}

//序列化块
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
//返学裂化
func DeserializeBlock(d []byte) *Block{
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil{
		panic("Block Deserialize Error!")
	}
	return &block
}

func (b *Block) HashTransactions() []byte{
	var txHashs = [][]byte{}
	var txHash [32]byte

	for _,value := range b.Transactions {
		txHashs = append(txHashs,value.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashs,[]byte{}))
	return txHash[:]
}

