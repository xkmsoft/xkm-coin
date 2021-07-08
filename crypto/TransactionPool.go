package crypto

import (
	"errors"
	"fmt"
)

var TransactionPool []Transaction = []Transaction{}

func GetTransactionById (id string) (bool, Transaction) {
	for i, _ := range TransactionPool {
		tx := TransactionPool[i]
		if tx.Id == id {
			return true, tx
		}
	}
	return false, Transaction{}
}

func AddToTransactionPool (transaction Transaction, unspentTxOuts []UnspentTxOut) error {
	if ValidateTransaction(&transaction, unspentTxOuts) != true {
		return errors.New("trying to add invalid tx to pool")
	}

	if !IsValidTxForPool(transaction, TransactionPool) {
		return errors.New("trying to add invalid tx to pool")
	}

	TransactionPool = append(TransactionPool, transaction)
	return nil
}

func UpdateTransactionPool () {
	unspentTxOuts := GetAllUnspentTxOuts()
	var newTransactionPool = []Transaction{}
	for i := range TransactionPool {
		tx := TransactionPool[i]
		if !HasTxIn(tx, unspentTxOuts) {
			newTransactionPool = append(newTransactionPool, tx)
		}
	}
	TransactionPool = newTransactionPool
}

func HasTxIn (transaction Transaction, unspentTxOuts []UnspentTxOut) bool {
	for i := range unspentTxOuts {
		txOut := unspentTxOuts[i]
		if txOut.TxOutId == transaction.Id {
			return true
		}
	}
	return false
}

func IsValidTxForPool(transaction Transaction, pool []Transaction) bool {
	txPoolIns := GetTxPoolIns(pool)
	for i := range transaction.TxIns {
		txIn := transaction.TxIns[i]
		if ContainsTxIn(txPoolIns, txIn) {
			fmt.Printf("Transaction pool already contains txIn\n")
			return false
		}
	}
	return true
}

func ContainsTxIn (txPoolIns []TxIn, txIn TxIn) bool {
	for i := range txPoolIns {
		txPoolIn := txPoolIns[i]
		if txPoolIn.TxOutIndex == txIn.TxOutIndex && txPoolIn.TxOutId == txIn.TxOutId {
			return true
		}
	}
	return false
}

func GetTxPoolIns (pool []Transaction) []TxIn {
	var txIns []TxIn
	for i := range pool {
		transaction := pool[i]
		for j := range transaction.TxIns {
			txIn := transaction.TxIns[j]
			txIns = append(txIns, txIn)
		}
	}
	return txIns
}