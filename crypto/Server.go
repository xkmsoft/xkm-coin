package crypto

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Params struct {
	Data []Transaction `json:"data"`
	Address string `json:"address"`
}

type NotFound struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	QUERY_LATEST = iota
	QUERY_ALL
	RESPONSE_BLOCKCHAIN
)

type Message struct {
	Id string `json:"id"`
	MessageType int `json:"message_type"`
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

var clients = make(map[*websocket.Conn] bool)

var broadcast = make(chan Message)

var upgradeWebSocket = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Blocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(BlockChain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func LatestBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(GetLatestBlock())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func Unspent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(GetAllUnspentTxOuts())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func GetTransactionPool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(TransactionPool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func HasTxOutId(txOutID string, txOuts []UnspentTxOut) int {
	for index, _ := range txOuts {
		txOut := txOuts[index]
		if txOut.TxOutId == txOutID {
			return index
		}
	}
	return -1
}

func HasTxOutIdWithAddress(txOutID string, address string, txOuts []UnspentTxOut) int {
	for index, _ := range txOuts {
		txOut := txOuts[index]
		if txOut.TxOutId == txOutID && txOut.Address == address {
			return index
		}
	}
	return -1
}

func RemoveElementFromSlice(slice []UnspentTxOut, index int) []UnspentTxOut {
	return append(slice[:index], slice[index+1:]...)
}

func GetAllUnspentTxOuts () []UnspentTxOut {
	var unspentTxOuts = []UnspentTxOut{}
	blockChain := GetBlockChain()
	for i, _ := range blockChain {
		block := blockChain[i]
		transactions := block.Data
		for j, _ := range transactions {
			transaction := transactions[j]
			txOuts := transaction.TxOuts
			for k, _ := range txOuts {
				txOut := txOuts[k]
				txIn := transaction.TxIns[0]
				index := HasTxOutIdWithAddress(txIn.TxOutId, txOut.Address, unspentTxOuts)
				if index != -1 {
					unspentTxOuts = RemoveElementFromSlice(unspentTxOuts, index)
				}
				unspentTxOuts = append(unspentTxOuts, UnspentTxOut{
					TxOutId:    transaction.Id,
					TxOutIndex: txIn.TxOutIndex,
					Address:    txOut.Address,
					Amount:     txOut.Amount,
				})
			}
		}
	}
	return unspentTxOuts
}

func GetBalanceOfUnspentTxOuts (txOuts []UnspentTxOut) int64 {
	var balance = int64(0)
	for i, _ := range txOuts {
		balance += txOuts[i].Amount
	}
	return balance
}

func GetUnspentTxOutsOfAddress(address string) []UnspentTxOut {
	var unspentTxOuts = []UnspentTxOut{}
	blockChain := GetBlockChain()
	for i, _ := range blockChain {
		block := blockChain[i]
		transactions := block.Data
		for j, _ := range transactions {
			transaction := transactions[j]
			txOuts := transaction.TxOuts
			for k, _ := range txOuts {
				txOut := txOuts[k]
				if txOut.Address == address {
					txIn := transaction.TxIns[0]
					index := HasTxOutId(txIn.TxOutId, unspentTxOuts)
					if index != -1 {
						unspentTxOuts = RemoveElementFromSlice(unspentTxOuts, index)
					}
					unspentTxOuts = append(unspentTxOuts, UnspentTxOut{
						TxOutId:    transaction.Id,
						TxOutIndex: txIn.TxOutIndex,
						Address:    txOut.Address,
						Amount:     txOut.Amount,
					})
				}
			}
		}
	}
	return unspentTxOuts
}

func Address(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["hash"]
	unspentTxOuts := GetUnspentTxOutsOfAddress(address)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(unspentTxOuts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type Wallet struct {
	Alias string `json:"alias"`
	Address string `json:"address"`
	Balance int64 `json:"balance"`
	UnspentTransactions int64 `json:"unspentTransactions"`
}

type StatusStruct struct {
	Circulation int64 `json:"circulation"`
	ChainSize int64 `json:"chainSize"`
	NumberOfWallets int64 `json:"numberOfWallets"`
	UnspentTxOuts int64 `json:"unspentTxOuts"`
	Wallets []Wallet `json:"wallets"`
}

func Status(w http.ResponseWriter, r *http.Request) {
	unSpentTxOuts := GetAllUnspentTxOuts()
	var balance = int64(0)
	duplicates := make(map[string]int)
	for i, _ := range unSpentTxOuts {
		txOut := unSpentTxOuts[i]
		balance += txOut.Amount
		_, exists := duplicates[txOut.Address]
		if !exists {
			duplicates[txOut.Address] = 1
		} else {
			duplicates[txOut.Address] += 1
		}
	}
	wallets := []Wallet{}
	for key, value := range duplicates {
		wallets = append(wallets, Wallet{
			Alias:               "",
			Address:             key,
			Balance:             GetBalanceOfUnspentTxOuts(GetUnspentTxOutsOfAddress(key)),
			UnspentTransactions: int64(value),
		})
	}
	status := StatusStruct{
		Circulation:     balance,
		ChainSize:       int64(len(GetBlockChain())),
		NumberOfWallets: int64(len(duplicates)),
		UnspentTxOuts:   int64(len(unSpentTxOuts)),
		Wallets:         wallets,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	blocks := GetBlockChain()
	var transaction *Transaction = nil
	for i := range blocks {
		for j := range blocks[i].Data {
			if blocks[i].Data[j].Id == id {
				transaction = &blocks[i].Data[j]
				break
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if transaction != nil {
		err := json.NewEncoder(w).Encode(transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	} else {
		err := json.NewEncoder(w).Encode(NotFound{Message: "Transaction not found"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func GetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	blocks := GetBlockChain()
	var block *Block = nil
	for i := range blocks {
		if blocks[i].Hash == hash {
			block = blocks[i]
			break
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if block != nil {
		err := json.NewEncoder(w).Encode(block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	} else {
		err := json.NewEncoder(w).Encode(NotFound{Message: "Block not found"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

type MineParams struct {
	Transactions []Transaction `json:"transactions"`
}

func MineBlock(w http.ResponseWriter, r *http.Request) {
	var params MineParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	added, block := GenerateNextBlock(params.Transactions)
	if added {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		err := json.NewEncoder(w).Encode(NotFound{Message: "Block is not valid"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

type SendTransactionStruct struct {
	Transaction Transaction `json:"transaction"`
	UnspentTxOuts []UnspentTxOut `json:"unspentTxOuts"`
}

func SendTransaction(w http.ResponseWriter, r *http.Request) {
	var params SendTransactionStruct
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = AddToTransactionPool(params.Transaction, params.UnspentTxOuts)
	if err != nil {
		err := json.NewEncoder(w).Encode(NotFound{Message: err.Error()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	found, tx := GetTransactionById(params.Transaction.Id)
	if found {
		err = json.NewEncoder(w).Encode(tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		err := json.NewEncoder(w).Encode(NotFound{Message: "Transaction could not be added to the transaction pool\n"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func HandleMessages()  {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				err := client.Close()
				if err != nil {
					return
				}
				delete(clients, client)
			}
		}
	}
}

func HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgradeWebSocket.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {

		}
	}(ws)

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}
