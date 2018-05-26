package Block

import (
	"github.com/boltdb/bolt"
	"log"
)

//数据库文件位置
const dbFile  = "blockchain.db"

//用于存放区块链信息的桶
//桶中结构
//键："l"   值：链中最后一个块的 hash
//键：区块的hash    值：区块的信息
const blocksBucket = "blocks"


//区块链结构
type BlockChain struct{
	tip []byte //保存最新块的hash
	DB *bolt.DB //数据库
}

//生成区块链迭代器
type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

//将区块添加到链尾
func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte
	//读取前一块hash
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//生成新区块 data和前一块hash
	newBlock := NewBlock(data,lastHash)
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//将新快的hash和新快序列化的内容保存早桶中
		err := b.Put(newBlock.Hash,newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		//再将新快的hash保存在l中
		err = b.Put([]byte("l"),newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		//新块的hash数据放入新块中
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//prevBlock := bc.Blocks[len(bc.Blocks)-1]
	//newBlock := NewBlock(data,prevBlock.Hash)
	//bc.Blocks = append(bc.Blocks,newBlock)
}

//生成创世块
func NewGenesisBlock() *Block {
    return NewBlock("Genesis Block", []byte{})
}

//生成一条新的区块链
func NewBlockChain() (blockChain *BlockChain){
	var tip []byte
	//打开数据库
	db,err := bolt.Open(dbFile,0600,nil)
	if err != nil {
		log.Panic(err)
	}
	//更新数据库内容
	db.Update(func(tx *bolt.Tx) error {
		//读取"blocks"桶中的二进制内容
		b := tx.Bucket([]byte(blocksBucket))
		if b== nil { //如果桶中没有数据 就生成创世块
			genesis := NewGenesisBlock()
			//创建一个桶
			b,err :=tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			//将创世块 也就是桶区块的hash放入键中，将创世块的二进制值放入值中
			err = b.Put(genesis.Hash,genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//桶中最后一块区块的hash放入"l"中
			err = b.Put([]byte("l"),genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			//将区块链的hash保存到区块链结构体中
			tip = genesis.Hash
		}else {
			//直接读取整个区块的hash
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	bc := BlockChain{tip,db}
	//blockChain = &BlockChain{[]*Block{NewGenesisBlock()}}
	return &bc
}

//生成的迭代器
//当前链的所有哈希
func (bc *BlockChain) Iterator() *BlockchainIterator{
	bci := &BlockchainIterator{bc.tip,bc.DB}
	return bci
}

//遍历
func (i *BlockchainIterator) Next() *Block{
	var block *Block
	//打开数据库以只读方式
	err := i.db.View(func(tx *bolt.Tx) error {
		//读取blocks桶中内容
		b := tx.Bucket([]byte(blocksBucket))
		//根据当前hash获得当前块的内容
		encodeBlock := b.Get(i.currentHash)
		//反序列化 还原当前块的信息
		block = DeserializeBlock(encodeBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//当指针前移
	i.currentHash = block.PrevHash
	return block
}