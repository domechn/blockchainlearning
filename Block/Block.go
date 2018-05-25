package Block

import (
	"time"

)

//定义一个区块链结构
type Block struct{
	Timestamp int64  //时间戳
	Data []byte  //携带数据
	PrevHash []byte  //前一块哈希值
	Hash []byte //哈希
	Nonce int //工作量证明
}


//生成一个区块并计算它的哈希值和计算量证明
func NewBlock(data string,prevBlockHash []byte) (block *Block) {
	//创建一个区块
	block = &Block{time.Now().Unix(),[]byte(data),prevBlockHash,[]byte{},0}
	//创建一个新的POW
	pow := NewproofOfWork(block)
	//生成哈希值和工作量证明
	nonce , hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return 
}