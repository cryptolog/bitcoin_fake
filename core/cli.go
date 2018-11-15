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
	fmt.Println("\"addblock -data BLOCK_DATA\"   -- add a block to the blockchain")
	fmt.Println("\"printchain\"   -- print all the blocks of the blockchain")
}

//校验命令输入合法性
func (cli *CLI) validateArgs(){
	if len(os.Args) < 2 {
		//未传递参数，输出帮助文档并退出
		cli.printUsage()
		os.Exit(1)
	}
}

// 往区块链上添加区块
func (cli *CLI) addBlock(data string){
	cli.Bc.AddBlock(data)
	fmt.Println("Success!")
}

//遍历输出区块链数据
func (cli *CLI) printChain(){
	bci := cli.Bc.Iterator()
	//迭代区块链并输出
	for{
		block := bci.Next()

		fmt.Printf("Prev.hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		//对该区块的PoW做一次验证
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		//当区块链空了便跳出循环
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

	//判断输入的命令(检查第二个参数，第一个为程序名)
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
		//未定义的命令，那就输出使用帮助
		cli.printUsage()
		//程序退出，返回值为1
		os.Exit(1)
	}

	//对应命令的调用代码
	if addBlockCmd.Parsed() {
		if *addBlockData == ""{
			//参数错误，输出使用帮助
			addBlockCmd.Usage()
			//程序退出，返回值为1
			os.Exit(1)
		}
		//调用添加区块到区块链的功能
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed(){
		//调用遍历区块链输出的功能
		cli.printChain()
	}
}