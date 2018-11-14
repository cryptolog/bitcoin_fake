package core

import (
	"fmt"
	"log"
	//bolt是一种开源的key-value数据储存库
	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"		//区块链数据存放文件
const blocksBucket = "blocks"		//区块数据存放‘桶’

// 创建一个 bolt “数据库”
type Blockchain struct {
	tip []byte
	Db *bolt.DB
}

//创建一个 bolt “数据库”
type BlockchainIterator struct {
	currentHash []byte
	Db *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(data, lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil{
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash

		return nil
	})
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip,bc.Db}
	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.Db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(blocksBucket))
		encodeBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodeBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
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

			//将区块数据存放到‘桶’
			err = b.Put(genesis.Hash, genesis.Serialize())
			//以上采用key,value存储，value为序列化的struct
			if err != nil {
				log.Panic(err)
			}

			//将该区块设为leader（第一）区块
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

	//不是很懂这里，将区块链放进内存并返回？
	bc := Blockchain{tip,db}
	return &bc
}