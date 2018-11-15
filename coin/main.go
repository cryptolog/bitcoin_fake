package main

import (
	"core"
)

func main() {
	bc := core.NewBlockchain()
	defer bc.Db.Close()

	cli := core.CLI{bc}	//command-line interface
	cli.Run()
}