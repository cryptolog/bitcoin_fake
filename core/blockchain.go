package core

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	//bolt是一种开源的key-value数据储存库
	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"		//区块链数据存放文件
const blocksBucket = "blocks"		//区块数据存放‘桶’
const genesisCoinbaseData = "Blank Data"

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
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
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
	newBlock := NewBlock(transactions, lastHash)
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
func NewBlockchain(address string) *Blockchain {
	//判断是否有区块链存在
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	//打开“区块链数据存放文件”，若失败则报错
	db,err := bolt.Open(dbFile,0600,nil)
	if err != nil{
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// CreateBlockchain 创建一个新的区块链数据库
// address 用来接收挖出创世块的奖励
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	//向“区块链数据存放文件”写入数据
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address,genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

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

// FindUnspentTransactions 找到未花费输出的交易
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// 如果交易输出被花费了
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// 如果该交易输出可以被解锁，即可被花费
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs 从 address 中找到至少 amount 的 UTXO
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}