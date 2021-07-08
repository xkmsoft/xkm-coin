package main

import (
	"chacoin/crypto"
	"fmt"
)

func main() {
	publicAddress := "02fbe9019062728e8fab7ac59b33d25c24ce9d393b49134f7a25da45a50f43faf9"
	public, err := crypto.GetPublicECDSAKeyFromCompressedAddress(publicAddress)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	fmt.Printf("Public Key: %+v\n", public)
}