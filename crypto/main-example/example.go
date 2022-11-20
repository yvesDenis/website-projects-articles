package main

import (
	"crypto-example/encryption"
	"fmt"
)

func main() {
	StringToEncrypt := "Encrypting this string"
	secretKey := "passphrase123456"
	// To encrypt the StringToEncrypt
	encText, err := encryption.Encrypt(StringToEncrypt, secretKey)
	if err != nil {
		fmt.Println("error encrypting your classified text: ", err)
	}
	fmt.Println(encText)
	// To decrypt the original StringToEncrypt
	decText, err := encryption.Decrypt("O5I/eiCXTEnOGa5fmRIiLmQ3L1O6Uw==", secretKey)
	if err != nil {
		fmt.Println("error decrypting your encrypted text: ", err)
	}
	fmt.Println(decText)
}
