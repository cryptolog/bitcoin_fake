package core

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64	//最大的随机数范围
)

const targetBits = 8		//计算的目标位数，该数越大难度越高

//工作量证明的结构
type ProofOfWork struct {
	block  *Block		//区块数据
	target *big.Int		//PoW计算目标
}

//创建一个目标target并传递
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256 - targetBits))
	//target左移位后与区块数据合并为一个结构体并返回
	pow := &ProofOfWork{b, target}
	return pow
}

//把区块的数据打包成一个data
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),	//计算难度
			IntToHex(int64(nonce)),		//自增随机数
		},
		[]byte{},
	)
	return data
}

//开始“挖矿”，计算某一符合条件的哈希
func (pow *ProofOfWork) Run() (int, []byte){
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n",pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		//计算打包后的数据并输出
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])
		//将哈希值转换为整数后与目标target进行对比
		if hashInt.Cmp(pow.target) == -1 {
			break
		}else{
			nonce++		//不满足目标则自增这个随机数，继续运算
		}
	}
	fmt.Print("\n\n")

	return nonce,hash[:]
}

//PoW有效性检验
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	//拿到区块数据和对应的随机数值
	data := pow.prepareData(pow.block.Nonce)
	//对区块数据进行sha256加密计算
	hash := sha256.Sum256(data)
	//把计算结果转换整数
	hashInt.SetBytes(hash[:])
	//与目标target进行比对（传进来的pow包含区块数据和target）
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}