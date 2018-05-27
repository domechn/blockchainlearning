package main

import "mybclearning/Block"

func main() {
	/*bc := Block.NewBlockChain()

	bc.AddBlock("this is my blockChain1")
	bc.AddBlock("this is my blockChain2")
	bc.AddBlock("this is my blockChain3")
	bc.AddBlock("this is my blockChain4")
	bc.AddBlock("this is my blockChain5")
	bc.AddBlock("this is my blockChain6")

	for _, block := range bc.Blocks {
		pow := Block.NewproofOfWork(block)
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("PoW: %s\n", pow.IsVaild())

		fmt.Println()
	}*/

	cli := Block.CLI{}
	cli.Run()
}
