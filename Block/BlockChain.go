package Block

import "fmt"

type BlockChain struct{
	blocks []*Block
}


func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data,prevBlock.Hash)
	bc.blocks = append(bc.blocks,newBlock)
}

func NewGenesisBlock() *Block {
    return NewBlock("Genesis Block", []byte{})
}

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