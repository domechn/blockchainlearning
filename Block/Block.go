package Block

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)

//定义一个区块链结构
type Block struct {
	Timestamp    int64          //时间戳
	Transactions []*Transaction //携带数据
	PrevHash     []byte         //前一块哈希值
	Hash         []byte         //哈希
	Nonce        int            //工作量证明
	Height		 int
}

//生成一个区块并计算它的哈希值和计算量证明
func NewBlock(transactions []*Transaction, prevBlockHash []byte , height int) (block *Block) {
	//创建一个区块
	block = &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0,height}
	//创建一个新的POW
	pow := NewproofOfWork(block)
	//生成哈希值和工作量证明
	nonce, hash := pow.Run()
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
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		panic("Block Deserialize Error!")
	}
	return &block
}

func (b *Block) HashTransactions() []byte {
	var txHashs = [][]byte{}

	for _, tx := range b.Transactions {
		txHashs = append(txHashs, tx.Serialize())
	}

	mTree := NewMerkleTree(txHashs)
	return mTree.RootNode.Data
}
