package core

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	Bc *Blockchain
}

//打印使用帮助文档
func (cli *CLI) printUsage(){
	fmt.Println("Usage:")
	fmt.Println("addblock -data BLOCK_DATA   -- add a block to the blockchain")
	fmt.Println("printchain   -- print all the blocks of the blockchain")
}

//校验命令输入合法性
func (cli *CLI) validateArgs(){
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// 往区块链上添加区块
func (cli *CLI) addBlock(data string){
	cli.Bc.AddBlock(data)
	fmt.Println("Success!")
}

func (cli *CLI) printChain(){
	bci := cli.Bc.Iterator()

	for{
		block := bci.Next()

		fmt.Printf("Prev.hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
}

//解析命令行参数并执行命令
func (cli *CLI) Run(){
	cli.validateArgs()
	//提供的可用命令
	addBlockCmd := flag.NewFlagSet("addblock",flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain",flag.ExitOnError)

	addBlockData := addBlockCmd.String("data","","Block data")
	//判断输入的命令
	switch  os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil{
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil{
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	//命令的执行代码
	if addBlockCmd.Parsed() {
		if *addBlockData == ""{
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed(){
		cli.printChain()
	}
}