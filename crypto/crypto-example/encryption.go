package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var bytes = []byte{57, 12, 34, 87, 32, 12, 90, 94, 69, 39, 77, 34, 52, 43, 69, 11}

func Encrypt(textToEncrypt string, secretKey string) (string, error) {
	//The key argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
	block, err := createNewCypher(secretKey)
	if err != nil {
		return "", err
	}

	plainText := []byte(textToEncrypt)
	//returns a BlockMode which encrypts in cipher block chaining mode, using the given Block.
	cfbEncrypter := cipher.NewCFBEncrypter(block, bytes)

	cipherText := make([]byte, len(plainText))
	cfbEncrypter.XORKeyStream(cipherText, plainText)

	return encode(cipherText), nil
}

func Decrypt(textToDecrypt string, secretKey string) (string, error) {

	block, err := createNewCypher(secretKey)
	if err != nil {
		return "", err
	}

	cipherText, err := decode(textToDecrypt)
	if err != nil {
		return "", err
	}
	cfbDecrypter := cipher.NewCFBDecrypter(block, bytes)

	plainText := make([]byte, len(cipherText))
	cfbDecrypter.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}

func createNewCypher(secretKey string) (cipher.Block, error) {
	return aes.NewCipher([]byte(secretKey))
}

func encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
func decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
