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

// 定义一个新区快并返回
func NewBlock(data string, prevBlockHash []byte) *Block {
	//声明一个区块（Block结构体）
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		}

	//进行一次PoW计算（挖矿）
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	//将计算得到的哈希和随机数保存为区块数据
	block.Hash = hash[:]
	block.Nonce = nonce
	
	return block
}

// 创世纪区块的创建
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
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

//反编码字节数组到区块数据
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}