package Block

import (
	"os"
	"fmt"
	"flag"
	"log"
	"strconv"
)

const usage = `
Usage:
  addblock -data BLOCK_DATA    add a block to the blockchain
  printchain                   print all the blocks of the blockchain
`

type CLI struct {
	BC *BlockChain
}

//打印提示操作
func (cli *CLI) printUsage() {
	fmt.Println(usage)
}
//判断用户输入是否合法 如果不合法打印提示信息 并退出系统
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

//打印整条区块链信息
func (cli *CLI) printChain() {
	bci := cli.BC.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewproofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.IsVaild()))
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		}
	}

}

//运行命令行界面
func (cli *CLI) Run() {
	//判断输入参数合法
	cli.validateArgs()
	//执行添加操作
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	//执行打印区块操作
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	//addblock -data "区块链data"
	addBlockData := addBlockCmd.String("data", "", "Block data")

	//判断输入内容 执行相应操作
	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	//获取到要添加的data  如果data不为空 则将data添加到区块链
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.BC.AddBlock(*addBlockData)
	}

	//打印整个区块链
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
