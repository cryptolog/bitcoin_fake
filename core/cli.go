package core

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {}

func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.Db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.Db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

//打印使用帮助文档
func (cli *CLI) printUsage(){
	fmt.Println("Usage:")
	fmt.Println("  printchain")
	fmt.Println("  getbalance -address ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT")
}

//校验命令输入合法性
func (cli *CLI) validateArgs(){
	if len(os.Args) < 2 {
		//未传递参数，输出帮助文档并退出
		cli.printUsage()
		os.Exit(1)
	}
}

//遍历输出区块链数据
func (cli *CLI) printChain(){
	bc := NewBlockchain("")
	defer bc.Db.Close()

	bci := bc.Iterator()
	//迭代区块链并输出
	for{
		block := bci.Next()

		fmt.Printf("Prev.hash: %x\n", block.PrevBlockHash)
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

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.AddBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

//解析命令行参数并执行命令
func (cli *CLI) Run(){
	cli.validateArgs()
	//提供的可用命令
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	//判断输入的命令(检查第二个参数，第一个为程序名)
	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		//未定义的命令，那就输出使用帮助
		cli.printUsage()
		//程序退出，返回值为1
		os.Exit(1)
	}

	//对应命令的调用代码
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed(){
		//调用遍历区块链输出的功能
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}