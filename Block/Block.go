package Block

import (
	"time"

)

type Block struct{
	Timestamp int64
	Data []byte
	PrevHash []byte
	Hash []byte
	Nonce int
}



func NewBlock(data string,prevBlockHash []byte) (block *Block) {
	block = &Block{time.Now().Unix(),[]byte(data),prevBlockHash,[]byte{},0}
	pow := NewproofOfWork(block)
	nonce , hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return 
}