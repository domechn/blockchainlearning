package Block

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"os"
	"fmt"
	"errors"
	"bytes"
	"crypto/ecdsa"
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

func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// 如果交易输出被花费了
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// 如果该交易输出可以被解锁，即可被花费
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
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
//将新的交易保存到数据库
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	for _,tx := range transactions{
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR:Invalid transaction")
		}
	}

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

//查找到未花费的输出（余额）
func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) FindUTXOs() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

//寻找所有能花费的输出  将输出累加起来达到需要交易的值时将输出的集合返回
func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
	//找到所有能花费的输出
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
	//不断循环累加 直到输出的值的和 大于需要付出的值
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIndex, out := range tx.Vout {
			if accumulated < amount && out.IsLockedWithKey(pubKeyHash){
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

//通过交易ID寻找对应的交易（为了找到本次输入关联的之前的交易）
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction,error){
	bci := bc.Iterator()
	for{
		block := bci.Next()

		for _ ,tx := range block.Transactions{
			if bytes.Compare(tx.ID,ID) == 0 || tx.IsCoinbase(){
				return *tx,nil
			}
		}

		if len(block.PrevHash) == 0{
			break
		}
	}
	return Transaction{},errors.New("Transaction is not found")
}

//对交易进行签名
/**
*	循环交易中的输入并找到输入对应的交易
*	并存入列表
*	将找到的交易用私钥进行加密
*/
func (bc *BlockChain) SignTransaction(tx *Transaction,privKey ecdsa.PrivateKey){
	prevTXs := make(map[string]Transaction)
	for _,vin := range tx.Vin{
		prevTX , err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)]=prevTX
	}
	tx.Sign(privKey,prevTXs)
}

//验证交易
/** 
*	循环交易中的输入并找到输入对应的交易
*	并存入列表
*	将找到的交易进行验证
*/
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool{
	prevTXs := make(map[string]Transaction)
	for _,vin := range tx.Vin{
		prevTX , err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)]=prevTX
	}
	return tx.Verify(prevTXs)
}