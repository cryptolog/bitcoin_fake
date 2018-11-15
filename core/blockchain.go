package core

import (
	"fmt"
	"log"
	//bolt是一种开源的key-value数据储存库
	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"		//区块链数据存放文件
const blocksBucket = "blocks"		//区块数据存放‘桶’

//  bolt “数据库”结构声明
type Blockchain struct {
	tip []byte
	Db *bolt.DB
}

// bolt “数据库”迭代器结构声明
type BlockchainIterator struct {
	currentHash []byte
	Db *bolt.DB
}

//申请添加一个新的区块
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		//打开blocksBucket = "blocks"的桶
		b := tx.Bucket([]byte(blocksBucket))
		//获取key为‘l(ast)’的区块哈希
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(data, lastHash)
	//生成一个新区块并序列化以便存入‘桶’
	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//更新bolt“数据库”，放入新区快的哈希值和编码后的数据
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil{
			log.Panic(err)
		}

		//将新区快的哈希key设置为‘l(ast)’
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
}

//迭代器，传递bolt的“数据库”
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip,bc.Db}
	return bci
}

//获取下一个区块位置
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.Db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(blocksBucket))
		//获取当前哈希的区块数据存放到encodeBlock
		encodeBlock := b.Get(i.currentHash)
		//将获取到的世数据反编码成区块数据
		block = DeserializeBlock(encodeBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//从反编码后的区块数据中获取上一区块哈希
	i.currentHash = block.PrevBlockHash

	return block
}

// 创建一条新的区块链
func NewBlockchain() *Blockchain {
	var tip []byte
	//打开“区块链数据存放文件”，若失败则报错
	db,err := bolt.Open(dbFile,0600,nil)
	if err != nil{
		log.Panic(err)
	}

	//向“区块链数据存放文件”写入数据
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//若“区块链数据存放文件”为空，新建创世区块
		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			//申请一个‘桶’
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			//将区块数据存放到‘桶’，采用key,value存储，value为编码的区块数据
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			//将该区块设为leader（第一）区块
			//其实称为leader或者last都行，最新一个也是放在最前面的
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil{
				log.Panic(err)
			}

			tip = genesis.Hash

		//若“区块链数据存放文件”不为空，获取leader区块
		}else{
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	//若写入数据失败，报错
	if err != nil{
		log.Panic(err)
	}

	//tip现在是最新区块的哈希，db为更新后的bolt“数据库”
	bc := Blockchain{tip,db}
	return &bc
}