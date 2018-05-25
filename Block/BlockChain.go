package Block

//区块链结构
type BlockChain struct{
	Blocks []*Block
}

//将区块添加到链尾
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(data,prevBlock.Hash)
	bc.Blocks = append(bc.Blocks,newBlock)
}

//生成创世块
func NewGenesisBlock() *Block {
    return NewBlock("Genesis Block", []byte{})
}

//生成区块链
func NewBlockChain() (blockChain *BlockChain){
	blockChain = &BlockChain{[]*Block{NewGenesisBlock()}}
	return 
}

/*func main() {
	bc := NewBlockChain()

	bc.AddBlock("this is my blockChain1")
	bc.AddBlock("this is my blockChain2")

	for _, block := range bc.blocks {
        fmt.Printf("Prev. hash: %x\n", block.PrevHash)
        fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
        fmt.Println()
    }

}*/