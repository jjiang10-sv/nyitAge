package main

// import (
// 	"bytes"
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"encoding/hex"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// )

// func main() {
// 	if len(os.Args) != 4 {
// 		fmt.Println("Usage: ./program <first> <second> <third>")
// 		return
// 	}

// 	first := os.Args[1]
// 	second := os.Args[2]
// 	third := os.Args[3]

// 	if len(first) != 21 {
// 		fmt.Println("First argument must be 21 characters long.")
// 		return
// 	}

// 	data := []byte(first)
// 	ciphertext, _ := hex.DecodeString(second)
// 	iv, _ := hex.DecodeString(third)

// 	keys, err := ioutil.ReadFile("./words.txt")
// 	if err != nil {
// 		fmt.Println("Error reading keys file:", err)
// 		return
// 	}

// 	for _, k := range string(keys) {
// 		k = k[:len(k)-1] // Remove newline character
// 		if len(k) <= 16 {
// 			key := string(k) + string(make([]byte, 16-len(k))) // Pad key to 16 bytes
// 			block, err := aes.NewCipher([]byte(key))
// 			if err != nil {
// 				fmt.Println("Error creating cipher:", err)
// 				continue
// 			}

// 			mode := cipher.NewCBCEncrypter(block, iv)
// 			paddedData := pad(data, aes.BlockSize)
// 			guess := make([]byte, len(paddedData))
// 			mode.CryptBlocks(guess, paddedData)

// 			if string(guess) == string(ciphertext) {
// 				fmt.Println("find the key:", key)
// 				return
// 			}
// 		}
// 	}

// 	fmt.Println("cannot find the key!")
// }

// func pad(data []byte, blockSize int) []byte {
// 	padding := blockSize - len(data)%blockSize
// 	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
// 	return append(data, padtext...)
// }
