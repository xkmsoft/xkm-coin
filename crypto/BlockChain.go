package crypto

import (
	"encoding/json"
	"fmt"
)

var GenesisTransaction = NewTransaction(
	"66ff05e7c66386297634cae4bd324e93be1b2ede6d4599e1de82361b42dc1807",
	[]TxIn{{
		TxOutId: "",
		TxOutIndex: 0,
		Signature: "",
	}},
	[]TxOut{{
		Address: "02fbe9019062728e8fab7ac59b33d25c24ce9d393b49134f7a25da45a50f43faf9",
		Amount: 100,
	}},
	)

var GenesisBlock = NewBlock(
	0,
	"46454b6c6f285e0d00437258b5a6543a0fcfadf278eb7e2b5cce151a383374a0",
	"",
	1,
	[]Transaction{*GenesisTransaction},
	0,
	0,
)

var BlockChain = []*Block{GenesisBlock}

func GetBlockChain() []*Block {
	return BlockChain
}

func GetLatestBlock() *Block {
	return BlockChain[len(BlockChain)-1]
}

func AddBlockToChain(block *Block) bool {
	if isValidBlock(block, GetLatestBlock()) {
		allUnspentTxOuts := GetAllUnspentTxOuts()
		_, err := ProcessTransactions(block.Data, allUnspentTxOuts, block.Index)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return false
		}
		BlockChain = append(BlockChain, block)
		UpdateTransactionPool()
		BroadcastBlock(block)
		return true
	}
	fmt.Printf("Not valid block\n")
	return false
}

func BroadcastBlock(block *Block) {
	var message Message
	timestamp := CurrentUnixTimestamp()
	message.Id = block.Hash
	message.Timestamp = timestamp
	message.MessageType = RESPONSE_BLOCKCHAIN
	out, err := json.Marshal(&block)
	if err != nil {
		message.Message = err.Error()
	}
	message.Message = string(out)
	broadcast <- message
}
