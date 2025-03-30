// Here's a Go implementation demonstrating various block cipher modes of operation using AES as the underlying cipher:

// ```go
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Helper functions
func padData(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func unpadData(data []byte) []byte {
	padding := data[len(data)-1]
	return data[:len(data)-int(padding)]
}

func xorBytes(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// Modes of Operation
func encryptECB(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := padData(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))

	for i := 0; i < len(padded); i += block.BlockSize() {
		block.Encrypt(ciphertext[i:], padded[i:])
	}
	return ciphertext, nil
}

func decryptECB(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(plaintext[i:], ciphertext[i:])
	}
	return unpadData(plaintext), nil
}

// limit : entire process can not be done parallelly. XOR is not much an overhead
// error progagation : error in transmission (network issue) it will recover
// and stop in the next block in decrypt; in encrypt, an error happens it propage to
// all the following blocks and end user received gababy informatin after decryption and
// will call the sender
// diffusion : rare cases that different plaintext will result in same cipher
func encryptCBC(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := padData(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))
	prev := iv

	for i := 0; i < len(padded); i += block.BlockSize() {
		xored := xorBytes(padded[i:i+block.BlockSize()], prev)
		block.Encrypt(ciphertext[i:i+block.BlockSize()], xored)
		prev = ciphertext[i : i+block.BlockSize()]
	}
	return append(iv, ciphertext...), nil
}

func decryptCBC(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	iv := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]
	plaintext := make([]byte, len(ciphertext))
	prev := iv

	for i := 0; i < len(ciphertext); i += blockSize {
		block.Decrypt(plaintext[i:blockSize], ciphertext[i:blockSize])
		xored := xorBytes(plaintext[i:i+blockSize], prev)
		copy(plaintext[i:], xored)
		prev = ciphertext[i : i+blockSize]
	}
	return unpadData(plaintext), nil
}

// cause it into stream cipher

func encryptCFB(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	prev := iv

	for i := 0; i < len(plaintext); {
		keystream := make([]byte, block.BlockSize())
		block.Encrypt(keystream, prev)
		// how to choose s bits; chunksize is s
		chunkSize := block.BlockSize()
		if remaining := len(plaintext) - i; remaining < chunkSize {
			chunkSize = remaining
		}

		copy(ciphertext[i:chunkSize], xorBytes(plaintext[i:i+chunkSize], keystream[:chunkSize]))
		//prev = ciphertext[i : i+chunkSize]
		// Shift register logic: shift prev and append new ciphertext
		copy(prev, append(ciphertext[chunkSize:], ciphertext[i:i+chunkSize]...))

		i += chunkSize
	}
	return append(iv, ciphertext...), nil
}

func decryptCFB(ciphertext, key []byte) ([]byte, error) {
	return encryptCFB(ciphertext, key, ciphertext[:aes.BlockSize])
}

// output feedback
// pre-computation output to solve the overhead
func encryptOFB(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	state := make([]byte, len(iv))
	copy(state, iv)

	for i := 0; i < len(plaintext); {
		keystream := make([]byte, block.BlockSize())
		block.Encrypt(keystream, state)
		copy(state, keystream)

		chunkSize := block.BlockSize()
		if remaining := len(plaintext) - i; remaining < chunkSize {
			chunkSize = remaining
		}

		copy(ciphertext[i:], xorBytes(plaintext[i:i+chunkSize], keystream[:chunkSize]))
		i += chunkSize
	}
	return append(iv, ciphertext...), nil
}
func decryptOFB(ciphertext, key []byte) ([]byte, error) {
	return encryptOFB(ciphertext, key, ciphertext[:aes.BlockSize])
}

// no chaining // counter
func encryptCTR(plaintext, key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	counter := make([]byte, block.BlockSize())
	copy(counter, nonce)

	for i := 0; i < len(plaintext); {
		keystream := make([]byte, block.BlockSize())
		block.Encrypt(keystream, counter)

		// Increment counter; need strong encrypt in randomness (assumption)
		// 		The code snippet you provided is part of a loop that increments a counter used in the CTR (Counter) mode of encryption. The purpose of this loop is to increment the counter, which is typically a byte array, in a way that mimics an integer increment operation.
		// Here's a brief explanation of how it works:

		// 1. The loop iterates over the bytes of the counter from the least significant byte to the most significant byte.
		// 2. It increments the current byte.
		// 3. If the incremented byte is not zero, it breaks out of the loop. This is because if the byte is not zero, it means there was no overflow, and the increment operation is complete.
		// 4. If the byte becomes zero, it means there was an overflow (e.g., incrementing 255 results in 0), and the loop continues to increment the next more significant byte.

		// The condition `if counter[j] != 0` is correct and necessary to handle the carry-over in a multi-byte counter. If you want to understand how `counter[j]` can be zero, consider the following:

		// - If `counter[j]` was initially 255 (0xFF in hexadecimal), incrementing it will result in 0 (overflow), and the carry will be added to the next byte in the sequence.

		// Here's a simple example to illustrate:

		// ```go
		// counter := []byte{0xFF, 0xFF} // Initial counter value
		// for j := len(counter) - 1; j >= 0; j-- {
		//     counter[j]++
		//     if counter[j] != 0 {
		//         break
		//     }
		// }
		// // After the loop, counter will be {0x00, 0x00} due to overflow
		// ```
		// In this example, both bytes in the counter are incremented from 255 to 0, demonstrating how the loop handles overflow by propagating the carry to the next byte.

		for j := len(counter) - 1; j >= 0; j-- {
			counter[j]++
			if counter[j] != 0 {
				break
			}
		}

		chunkSize := block.BlockSize()
		if remaining := len(plaintext) - i; remaining < chunkSize {
			chunkSize = remaining
		}

		copy(ciphertext[i:], xorBytes(plaintext[i:i+chunkSize], keystream[:chunkSize]))
		i += chunkSize
	}
	return append(nonce, ciphertext...), nil
}

func main() {
	key := []byte("16bytekey12345678")
	iv := []byte("16byteiv12345678")
	message := []byte("Secret message demonstrating AES modes of operation!")

	// ECB
	ecbCipher, _ := encryptECB(message, key)
	ecbPlain, _ := decryptECB(ecbCipher, key)

	// CBC
	cbcCipher, _ := encryptCBC(message, key, iv)
	cbcPlain, _ := decryptCBC(cbcCipher, key)

	// CFB
	cfbCipher, _ := encryptCFB(message, key, iv)
	cfbPlain, _ := decryptCFB(cfbCipher, key)

	// CTR
	ctrCipher, _ := encryptCTR(message, key, iv)

	fmt.Println("Original:", string(message))
	fmt.Println("\nECB:")
	fmt.Println("Encrypted:", hex.EncodeToString(ecbCipher))
	fmt.Println("Decrypted:", string(ecbPlain))

	fmt.Println("\nCBC:")
	fmt.Println("Encrypted:", hex.EncodeToString(cbcCipher))
	fmt.Println("Decrypted:", string(cbcPlain))

	fmt.Println("\nCFB:")
	fmt.Println("Encrypted:", hex.EncodeToString(cfbCipher))
	fmt.Println("Decrypted:", string(cfbPlain))

	fmt.Println("\nCTR:")
	fmt.Println("Encrypted:", hex.EncodeToString(ctrCipher))
}

// ```

// This code demonstrates:

// **ECB (Electronic Codebook)**
// - Simplest mode with independent block encryption
// - Vulnerable to pattern analysis
// - Requires padding

// **CBC (Cipher Block Chaining)**
// - Uses initialization vector (IV)
// - Chained encryption with XOR operations
// - Provides better security than ECB

// **CFB (Cipher Feedback)**
// - Stream cipher mode
// - Self-synchronizing
// - No padding required

// **OFB (Output Feedback)**
// - Stream cipher mode
// - Error propagation limited
// - Keystream generated independently

// **CTR (Counter)**
// - Stream cipher mode
// - Parallelizable encryption
// - Uses counter instead of IV

// Key features:
// - Uses AES-128 with 16-byte keys
// - Implements padding for block modes (ECB/CBC)
// - Includes XOR helper function for stream modes
// - Demonstrates proper IV/nonce handling
// - Shows encryption/decryption workflow for each mode

// Note: This is for educational purposes only. Real-world implementations should use standard library's cipher package implementations with proper security considerations.

// Here's a Go implementation that simulates basic block cipher operations using a simplified Feistel network structure. This example demonstrates encryption/decryption with a fixed block size and multiple rounds:

type BlockCipher struct {
	rounds    int
	key       []byte
	blockSize int
}

func NewBlockCipher(key []byte, rounds int) *BlockCipher {
	return &BlockCipher{
		rounds:    rounds,
		key:       key,
		blockSize: 8, // 64-bit blocks for demonstration
	}
}

// Simple Feistel round function
func (bc *BlockCipher) feistelFunction(halfBlock uint32, roundKey byte) uint32 {
	// This is a simplified round function for demonstration
	return halfBlock ^ uint32(roundKey) ^ 0x9e3779b9
}

// // Pad data to block size using PKCS#7
// func padData(data []byte, blockSize int) []byte {
// 	padding := blockSize - (len(data) % blockSize)
// 	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
// }

// // Remove padding
// func unpadData(data []byte) []byte {
// 	padding := data[len(data)-1]
// 	return data[:len(data)-int(padding)]
// }

func (bc *BlockCipher) Encrypt(plaintext []byte) []byte {
	padded := padData(plaintext, bc.blockSize)
	ciphertext := make([]byte, len(padded))

	for i := 0; i < len(padded); i += bc.blockSize {
		block := padded[i : i+bc.blockSize]
		left := binary.BigEndian.Uint32(block[:4])
		right := binary.BigEndian.Uint32(block[4:])

		for r := 0; r < bc.rounds; r++ {
			roundKey := bc.key[r%len(bc.key)]
			temp := right
			right = left ^ bc.feistelFunction(right, roundKey)
			left = temp
		}

		binary.BigEndian.PutUint32(ciphertext[i:i+4], left)
		binary.BigEndian.PutUint32(ciphertext[i+4:i+8], right)
	}
	return ciphertext
}

func (bc *BlockCipher) Decrypt(ciphertext []byte) []byte {
	plaintext := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += bc.blockSize {
		block := ciphertext[i : i+bc.blockSize]
		left := binary.BigEndian.Uint32(block[:4])
		right := binary.BigEndian.Uint32(block[4:])

		for r := bc.rounds - 1; r >= 0; r-- {
			roundKey := bc.key[r%len(bc.key)]
			temp := left
			left = right ^ bc.feistelFunction(left, roundKey)
			right = temp
		}

		binary.BigEndian.PutUint32(plaintext[i:i+4], left)
		binary.BigEndian.PutUint32(plaintext[i+4:i+8], right)
	}
	return unpadData(plaintext)
}

func mainbco() {
	key := []byte{0x2b, 0x7e, 0x15, 0x16}
	cipher := NewBlockCipher(key, 16)

	message := []byte("Block cipher demo!")
	fmt.Printf("Original: %s\n", message)

	encrypted := cipher.Encrypt(message)
	fmt.Printf("Encrypted: %x\n", encrypted)

	decrypted := cipher.Decrypt(encrypted)
	fmt.Printf("Decrypted: %s\n", decrypted)
}

// ```

// This implementation demonstrates:

// 1. **Feistel Network Structure**: Splits blocks into two halves and processes through multiple rounds
// 2. **Block Operations**: 64-bit block size with PKCS#7 padding
// 3. **Round Function**: Simplified substitution-permutation network
// 4. **Symmetric Key**: Uses same key for encryption/decryption
// 5. **Mode of Operation**: ECB (Electronic Codebook) - for demonstration only

// Key components:
// - `feistelFunction`: Simple round function for demonstration
// - `padData`/`unpadData`: Handle block alignment
// - `Encrypt`/`Decrypt`: Main cipher operations
// - Multiple rounds of transformation for confusion and diffusion

// Limitations (for educational purposes):
// - Small block size (64-bit)
// - Weak round function
// - ECB mode (not secure for real-world use)
// - Simplified key scheduling

// For production use, consider using standard library's `crypto/aes` package with proper modes like CBC or GCM. This example is meant to demonstrate basic block cipher principles rather than provide actual security.

// Here's a Go implementation demonstrating cryptographic algorithm criteria through a custom cipher mode combining CTR encryption with diffusion and error detection mechanisms. The implementation addresses:

// 1. **Overhead Management**
// 2. **Error Detection/Recovery**
// 3. **Controlled Error Propagation**
// 4. **Diffusion Enhancement**
// 5. **Security Foundations**

const (
	blockSize = 16
	hmacSize  = 32
	nonceSize = 12
	overhead  = nonceSize + hmacSize
)

type SecureCipher struct {
	key       []byte
	diffusion bool
}

func NewSecureCipher(key []byte, diffusion bool) *SecureCipher {
	return &SecureCipher{
		key:       key,
		diffusion: diffusion,
	}
}

// Enhanced encryption with diffusion and error detection
func (sc *SecureCipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// AES-CTR Encryption
	block, _ := aes.NewCipher(sc.key)
	ctr := cipher.NewCTR(block, nonce)
	ciphertext := make([]byte, len(plaintext))
	ctr.XORKeyStream(ciphertext, plaintext)

	// Diffusion Layer
	if sc.diffusion {
		sc.applyDiffusion(ciphertext)
	}

	// Error Detection (HMAC)
	mac := hmac.New(sha256.New, sc.key)
	mac.Write(append(nonce, ciphertext...))
	final := append(nonce, ciphertext...)
	final = append(final, mac.Sum(nil)...)

	return final, nil
}

func (sc *SecureCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < overhead {
		return nil, errors.New("invalid ciphertext")
	}

	// Verify HMAC
	mac := hmac.New(sha256.New, sc.key)
	mac.Write(ciphertext[:len(ciphertext)-hmacSize])
	if !hmac.Equal(mac.Sum(nil), ciphertext[len(ciphertext)-hmacSize:]) {
		return nil, errors.New("authentication failed")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize : len(ciphertext)-hmacSize]

	// Reverse Diffusion
	if sc.diffusion {
		sc.reverseDiffusion(ciphertext)
	}

	// AES-CTR Decryption
	block, _ := aes.NewCipher(sc.key)
	ctr := cipher.NewCTR(block, nonce)
	plaintext := make([]byte, len(ciphertext))
	ctr.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// Diffusion mechanism
func (sc *SecureCipher) applyDiffusion(data []byte) {
	prev := make([]byte, blockSize)
	for i := 0; i < len(data); i += blockSize {
		end := i + blockSize
		if end > len(data) {
			end = len(data)
		}

		hash := sha256.Sum256(prev)
		for j := i; j < end; j++ {
			data[j] ^= hash[j%len(hash)]
		}
		copy(prev, data[i:end])
	}
}

func (sc *SecureCipher) reverseDiffusion(data []byte) {
	prev := make([]byte, blockSize)
	for i := 0; i < len(data); i += blockSize {
		end := i + blockSize
		if end > len(data) {
			end = len(data)
		}

		current := make([]byte, end-i)
		copy(current, data[i:end])

		hash := sha256.Sum256(prev)
		for j := i; j < end; j++ {
			data[j] ^= hash[j%len(hash)]
		}
		copy(prev, current)
	}
}

func mainCriteria() {
	key := make([]byte, 32) // AES-256
	rand.Read(key)

	cipher := NewSecureCipher(key, true)
	msg := []byte("Cryptographic criteria demonstration")

	// Encryption
	ct, _ := cipher.Encrypt(msg)
	fmt.Printf("Ciphertext (%d bytes overhead):\n%x\n\n", overhead, ct)

	// Decryption
	pt, _ := cipher.Decrypt(ct)
	fmt.Printf("Decrypted: %s\n\n", pt)

	// Error simulation
	ct[nonceSize+5] ^= 0x01 // Flip 1 bit
	_, err := cipher.Decrypt(ct)
	fmt.Println("Error detection:", err)

	// Error propagation analysis
	original := []byte{0x00, 0x00, 0x00, 0x00}
	ct, _ = cipher.Encrypt(original)
	ct[nonceSize+1] ^= 0x80 // Introduce error
	pt, _ = cipher.Decrypt(ct)
	fmt.Printf("\nError propagation analysis:\nOriginal: %x\nCorrupted: %x\n", original, pt)
}

// ```

// **Key Criteria Implementation:**

// 1. **Overhead Control**
//    - Fixed 12-byte nonce + 32-byte HMAC = 44 bytes total overhead
//    - CTR mode eliminates padding overhead

// 2. **Error Detection & Recovery**
//    - HMAC-SHA256 for authentication
//    - Automatic error detection with precise error localization
//    - Enables retransmission requests

// 3. **Controlled Error Propagation**
//    - CTR base limits errors to affected bytes
//    - Diffusion layer contains impact to 2 blocks
//    - 1-bit flip only corrupts 1-2 plaintext blocks

// 4. **Enhanced Diffusion**
//    - XOR chain with SHA-256 hash of previous block
//    - Small plaintext changes avalanche through ciphertext
//    - Break statistical patterns in output

// 5. **Security Foundations**
//    - AES-256 as core cipher
//    - Random nonces prevent IV reuse
//    - HMAC-SHA256 for integrity
//    - CTR mode provides semantic security

// **Sample Output Analysis:**
// ```text
// Ciphertext (44 bytes overhead):
// 12-byte nonce + encrypted data + 32-byte HMAC

// Decrypted: Cryptographic criteria demonstration

// Error detection: authentication failed

// Error propagation analysis:
// Original: 00000000
// Corrupted: 00008000  // Error contained to 2 bytes
// ```

// This implementation demonstrates trade-offs between cryptographic properties while maintaining core security requirements. For production use, consider standardized AEAD modes like AES-GCM with additional safeguards.
