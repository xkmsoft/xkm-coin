package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	BlockGenerationInterval      = 10
	DifficultyAdjustmentInterval = 10
)

type Block struct {
	Index        int64  `json:"index"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previousHash"`
	Timestamp    int64  `json:"timestamp"`
	Data         []Transaction `json:"data"`
	Difficulty   int    `json:"difficulty"`
	Nonce        uint32 `json:"nonce"`
}

func NewBlock(index int64, hash string, previousHash string, timestamp int64, data []Transaction, difficulty int, nonce uint32) *Block {
	block := Block{
		Index:        index,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		Data:         data,
		Difficulty:   difficulty,
		Nonce:        nonce,
	}
	return &block
}

func GetDifficulty(chain []*Block) int {
	latest := chain[len(chain)-1]
	if latest.Index%DifficultyAdjustmentInterval == 0 && latest.Index != 0 {
		return getAdjustedDifficulty(latest, chain)
	}
	return latest.Difficulty
}

func getAdjustedDifficulty(latest *Block, chain []*Block) int {
	previousAdjustmentBlock := chain[len(chain)-DifficultyAdjustmentInterval]
	timeExpected := int64(BlockGenerationInterval * DifficultyAdjustmentInterval)
	timeTaken := latest.Timestamp - previousAdjustmentBlock.Timestamp
	if timeTaken < timeExpected/2 {
		return previousAdjustmentBlock.Difficulty + 1
	} else if timeTaken > timeExpected*2 {
		if previousAdjustmentBlock.Difficulty > 0 {
			return previousAdjustmentBlock.Difficulty - 1
		}
		return 0
	} else {
		return previousAdjustmentBlock.Difficulty
	}
}

func GenerateNextBlock(data []Transaction) (bool, *Block) {
	previousBlock := GetLatestBlock()
	nextIndex := previousBlock.Index + 1
	difficulty := GetDifficulty(GetBlockChain())
	nextTimeStamp := CurrentUnixTimestamp()
	block := FindBlock(nextIndex, previousBlock.Hash, nextTimeStamp, data, difficulty)
	if isValidBlock(block, previousBlock) {
		added := AddBlockToChain(block)
		return added, block

	}
	return false, block
}

func FindBlock(index int64, previousHash string, timestamp int64, data []Transaction, difficulty int) *Block {
	var nonce uint32 = 0
	for {
		hash := CalculateHash(index, previousHash, timestamp, data, difficulty, nonce)
		if HashMatchesDifficulty(hash, difficulty) {
			block := NewBlock(index, hash, previousHash, timestamp, data, difficulty, nonce)
			fmt.Printf("New block found! %+v\n", block)
			return block
		}
		nonce++
	}
}

func HexToBinary(h string) string {
	decoded, _ := hex.DecodeString(h)
	var builder strings.Builder
	for _, b := range decoded {
		builder.WriteString(fmt.Sprintf("%08b", b))
	}
	return builder.String()
}

func HashMatchesDifficulty(hash string, difficulty int) bool {
	hashInBinary := HexToBinary(hash)
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hashInBinary, prefix)
}

func CurrentUnixTimestamp() int64 {
	return time.Now().UTC().Unix()
}

func CalculateHashForBlock(block *Block) string {
	return CalculateHash(block.Index, block.PreviousHash, block.Timestamp, block.Data, block.Difficulty, block.Nonce)
}

func CalculateHash(index int64, previousHash string, timestamp int64, data []Transaction, difficulty int, nonce uint32) string {
	str := fmt.Sprintf("%d%s%d%v%d%d", index, previousHash, timestamp, data, difficulty, nonce)
	return HashString(str)
}

func HashString(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

func isValidBlock(newBlock *Block, previousBlock *Block) bool {
	if previousBlock.Index+1 != newBlock.Index {
		fmt.Printf("Index mismatch")
		return false
	}
	if previousBlock.Hash != newBlock.PreviousHash {
		fmt.Printf("Previous hash mismatch")
		return false
	}
	if CalculateHashForBlock(newBlock) != newBlock.Hash {
		fmt.Printf("Hash mismatch")
		return false
	}
	if !isValidTimestamp(newBlock, previousBlock) {
		fmt.Printf("Non valid timestamp")
		return false
	}
	if !HasValidHash(newBlock) {
		fmt.Printf("Non valid hash")
		return false
	}
	return true
}

func HasValidHash(block *Block) bool {
	if !HasMatchesBlockContent(block) {
		return false
	}
	if !HashMatchesDifficulty(block.Hash, block.Difficulty) {
		return false
	}
	return true
}

func HasMatchesBlockContent(block *Block) bool {
	hash := CalculateHashForBlock(block)
	return block.Hash == hash
}

func isValidTimestamp(newBlock *Block, previousBlock *Block) bool {
	return (previousBlock.Timestamp-60 < newBlock.Timestamp) && newBlock.Timestamp-60 < CurrentUnixTimestamp()
}
