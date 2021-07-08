package main

import (
	"chacoin/crypto"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/blocks", crypto.Blocks).Methods("GET")
	router.HandleFunc("/api/status", crypto.Status).Methods("GET")
	router.HandleFunc("/api/LatestBlock", crypto.LatestBlock).Methods("GET")
	router.HandleFunc("/api/unspent", crypto.Unspent).Methods("GET")
	router.HandleFunc("/api/block/{hash}", crypto.GetBlock).Methods("GET")
	router.HandleFunc("/api/address/{hash}", crypto.Address).Methods("GET")
	router.HandleFunc("/api/transaction/{id}", crypto.GetTransaction).Methods("GET")
	router.HandleFunc("/api/transactionPool", crypto.GetTransactionPool).Methods("GET")
	router.HandleFunc("/api/sendTransaction", crypto.SendTransaction).Methods("POST")
	router.HandleFunc("/api/mine", crypto.MineBlock).Methods("POST")
	router.HandleFunc("/ws", crypto.HandleWSConnections)
	go crypto.HandleMessages()
	log.Fatal(http.ListenAndServe(":3000", router))
}