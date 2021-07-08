package crypto

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CoinBaseAmount      = 100
)

type UnspentTxOut struct {
	TxOutId    string `json:"txOutId"`
	TxOutIndex int64  `json:"txOutIndex"`
	Address    string `json:"address"`
	Amount     int64  `json:"amount"`
}

func NewUnspentTxOut(txOutId string, txOutIndex int64, address string, amount int64) *UnspentTxOut {
	txOut := UnspentTxOut{
		TxOutId:    txOutId,
		TxOutIndex: txOutIndex,
		Address:    address,
		Amount:     amount,
	}
	return &txOut
}

type TxIn struct {
	TxOutId    string `json:"txOutId"`
	TxOutIndex int64 `json:"txOutIndex"`
	Signature  string `json:"signature"`
}

type TxOut struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type Transaction struct {
	Id     string  `json:"id"`
	TxIns  []TxIn  `json:"txIns"`
	TxOuts []TxOut `json:"txOuts"`
}

func NewTransaction(id string, txIns []TxIn, txOuts []TxOut) *Transaction {
	transaction := Transaction{
		Id:     id,
		TxIns:  txIns,
		TxOuts: txOuts,
	}
	return &transaction
}

func GetTransactionId (transaction *Transaction) string {
	var inBuilder, outBuilder strings.Builder
	for i := range transaction.TxIns {
		inBuilder.WriteString(fmt.Sprintf("%s%d", transaction.TxIns[i].TxOutId, transaction.TxIns[i].TxOutIndex))
	}
	for j := range transaction.TxOuts {
		outBuilder.WriteString(fmt.Sprintf("%s%d", transaction.TxOuts[j].Address, transaction.TxOuts[j].Amount))
	}
	combined := fmt.Sprintf("%s%s", inBuilder.String(), outBuilder.String())
	hashed := HashString(combined)
	return hashed
}

func ValidateTransaction (transaction *Transaction, unspentTxOuts []UnspentTxOut) bool {
	if GetTransactionId(transaction) != transaction.Id {
		fmt.Printf("Transaction ids do not match\n")
		return false
	}
	// Validation of TxIns
	for i := range transaction.TxIns {
		validated := ValidateTxIn(&transaction.TxIns[i], transaction, unspentTxOuts)
		if !validated {
			fmt.Printf("txin is not valid: %+v\n", &transaction.TxIns[i])
			return false
		}
	}

	var totalTxInValues int64
	var totalTxOutValues int64

	for i := range transaction.TxIns {
		txOut := FindReferencedTxOut(&transaction.TxIns[i], unspentTxOuts)
		if txOut != nil {
			totalTxInValues += txOut.Amount
		}
	}

	for i := range transaction.TxOuts {
		totalTxOutValues += transaction.TxOuts[i].Amount
	}

	if totalTxInValues != totalTxOutValues {
		fmt.Printf("Total txIn and txOut valued do not match. TxIns: %d TxOuts: %d\n", totalTxInValues, totalTxOutValues)
		return false
	}

	return true
}

func FindReferencedTxOut (txIn *TxIn, unspentTxOuts []UnspentTxOut) *UnspentTxOut {
	for i := range unspentTxOuts {
		if unspentTxOuts[i].TxOutId == txIn.TxOutId && unspentTxOuts[i].TxOutIndex == txIn.TxOutIndex {
			return &unspentTxOuts[i]
		}
	}
	return nil
}

func ValidateTxIn (txIn *TxIn, transaction *Transaction, unspentTxOuts []UnspentTxOut) bool {
	referencedTxOut := FindReferencedTxOut(txIn, unspentTxOuts)
	if referencedTxOut == nil {
		fmt.Printf("Referenced txOut not found")
		return false
	}
	address := referencedTxOut.Address
	publicKey, err := GetPublicECDSAKeyFromCompressedAddress(address)
	if err != nil {
		fmt.Printf("Public key could not be derived from address: %s\n", err.Error())
		return false
	}
	validated, err := VerifyECDSASignature(publicKey, transaction.Id, txIn.Signature)
	if err != nil {
		fmt.Printf("Signature could not be verified: %s\n", err.Error())
		return false
	}
	return validated
}

func ProcessTransactions (transactions []Transaction, unspentTxOuts []UnspentTxOut, blockIndex int64) (bool, error) {
	if !ValidateBlockTransactions(transactions, unspentTxOuts, blockIndex) {
		return false, errors.New("invalid block transactions\n")
	}
	return true, nil
}

func ValidateBlockTransactions (transactions []Transaction, unspentTxOuts []UnspentTxOut, blockIndex int64) bool {
	coinBaseTx := transactions[0]
	if !ValidateCoinBaseTx(&coinBaseTx, blockIndex) {
		fmt.Printf("Invalid coinbase transaction: %+v\n", coinBaseTx)
		return false
	}
	var txIns = []TxIn{}
	for i := range transactions {
		tx := transactions[i]
		for j := range tx.TxIns {
			txIns = append(txIns, tx.TxIns[j])
		}
	}

	if HasDuplicates(txIns) {
		return false
	}

	normalTransactions := transactions[1:]
	for _, tx := range normalTransactions {
		if !ValidateTransaction(&tx, unspentTxOuts) {
			return false
		}
	}
	return true
}

func HasDuplicates (txIns []TxIn) bool {
	duplicates := make(map[string]int)
	for _, txIn := range txIns {
		combined := fmt.Sprintf("%s%d", txIn.TxOutId, txIn.TxOutIndex)
		_, exists := duplicates[combined]
		if exists {
			return true
		} else {
			duplicates[combined] = 1
		}
	}
	return false
}

func ValidateCoinBaseTx (transaction *Transaction, blockIndex int64) bool {
	if transaction == nil {
		fmt.Printf("Transaction is nil\n")
		return false
	}
	if GetTransactionId(transaction) != transaction.Id {
		fmt.Printf("Transaction ids do not match\n")
		return false
	}
	if len(transaction.TxIns) != 1 {
		fmt.Printf("one txIn must be specified in the coinbase transaction\n")
		return false
	}
	if transaction.TxIns[0].TxOutIndex != blockIndex {
		fmt.Printf("the txin signature in coinbase tx must be the block height\n")
		return false
	}
	if transaction.TxOuts[0].Amount != CoinBaseAmount {
		fmt.Printf("Invalid coinbase amount in coinbase transaction\n")
		return false
	}
	return true
}