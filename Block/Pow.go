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

//定义POW 由当前区块和需要计算的值
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

//创建一个POW
func NewproofOfWork(b *Block) (pow *ProofOfWork) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow = &ProofOfWork{b, target}
	return
}

//执行计算（挖矿）
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	//打印当前块的数据
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	//设置边界 计算以防越界
	for nonce < math.MaxInt64 {
		//计算出需要求hash的数据
		data := pow.prepareData(nonce)
		//求哈希值
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		//如果算出来的hash比约定的小就返回hash值和工作量
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

//检验区块是否合法 如果当前块的hash小于约定值 说明合法
func (pow *ProofOfWork) IsVaild() bool{
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}

//将工作量和区块数据打包
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
