package Block

import (
	"math/big"
	"bytes"
	"fmt"
	"math"
	"crypto/sha256"
	"encoding/binary"
)

const targetBits = 24

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewproofOfWork(b *Block) (pow *ProofOfWork) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow = &ProofOfWork{b, target}
	return
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce ++
		}

	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func (pow *ProofOfWork) IsVaild() bool{
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}

func (pow *ProofOfWork) prepareData(nonce int) (data []byte) {
	data = bytes.Join([][]byte{
		pow.block.PrevHash,
		pow.block.Data,
		Int64ToBytes(pow.block.Timestamp),
		Int64ToBytes(int64(targetBits)),
		Int64ToBytes(int64(nonce)),
	}, []byte{})
	return
}
