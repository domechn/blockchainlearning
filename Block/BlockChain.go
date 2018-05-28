package Block

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"os"
	"fmt"
)

//数据库文件位置
const dbFile = "blockchain.db"

//用于存放区块链信息的桶
//桶中结构
//键："l"   值：链中最后一个块的 hash
//键：区块的hash    值：区块的信息
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

//区块链结构
type BlockChain struct {
	tip []byte   //保存最新块的hash
	DB  *bolt.DB //数据库
}

//生成区块链迭代器
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//将区块添加到链尾
func (bc *BlockChain) AddBlock(transactions []*Transaction) {
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
	newBlock := NewBlock(transactions, lastHash)
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//将新快的hash和新快序列化的内容保存早桶中
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		//再将新快的hash保存在l中
		err = b.Put([]byte("l"), newBlock.Hash)
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
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

//生成一条新的区块链
func CreateBlockchain(address string) (blockChain *BlockChain) {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	//打开数据库
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	//更新数据库内容
	err = db.Update(func(tx *bolt.Tx) error {
		//读取"blocks"桶中的二进制内容
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := BlockChain{tip, db}
	fmt.Printf("Hash: %x\n", tip)
	//blockChain = &BlockChain{[]*Block{NewGenesisBlock()}}
	return &bc
}

func NewBlockchain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	//打开数据库
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := &BlockChain{tip, db}
	return bc
}

//生成的迭代器
//当前链的所有哈希
func (bc *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.DB}
	return bci
}

//遍历
func (i *BlockchainIterator) Next() *Block {
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

func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, transaction := range block.Transactions {
			txID := hex.EncodeToString(transaction.ID)
		Outputs:
			for outIdx, out := range transaction.Vout {
				if spentTXOs[txID] != nil {
					for _, spendOut := range spentTXOs[txID] {
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				if (out.CanBeUnlockedWith(address)) {
					unspentTXs = append(unspentTXs, *transaction)
				}
			}

			if transaction.IsCoinbase() == false{
				for _, in := range transaction.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(transactions,lastHash)
	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash,newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"),newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
}

func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIndex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIndex)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}
