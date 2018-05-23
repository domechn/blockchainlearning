package Pow


import (
	"math/big"
	"bytes"
)

const targetBits = 24

type ProofOfWork struct{
	block *Block
	target *big.Int
}


func NewproofOfWork(b *Block) (pow *ProofOfWork) {
	target := big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))
	pow = &{b,target}
}

func (pow *ProofOfWork) prepareData(nonce int) (data []byte) {
	data = bytes.Join([][]byte{
			pow.block.PrevHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce))
		},[]byte{})
	return 
}


