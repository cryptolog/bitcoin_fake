package core

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// 区块结构的声明
type Block struct {
	Timestamp     int64  // 区块创建的时间戳
	Data          []byte // 区块中包含的数据
	PrevBlockHash []byte // 前一个区块的哈希
	Hash          []byte // 当前区块的哈希，可用于校验区块数据有效性
	Nonce         int    //用于验证工作量证明的随机数
}

//编码区块数据为字节数组
func (b *Block) Serialize() []byte{
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil{
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

// 新区快的创建
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		}
	pow :=NewProofOfWork(block)
	nonce, hash := pow.Run()
	
	block.Hash = hash[:]
	block.Nonce = nonce
	
	return block
}

/* 计算并设置哈希值
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}
*/

// 创世纪区块的创建
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
