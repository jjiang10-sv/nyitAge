package main

import (
	"fmt"
	"testing"
)

func TestDecryptSim(t *testing.T) {
	//decrypTable := getConstructionTable()
	decrypTable := getDecryptTableComp()
	// limited to uppercase
	ciphertext := "JHJKHasad"
	plaintext := decrypt(ciphertext, decrypTable)
	fmt.Println(plaintext)
}

func TestModInverse(t *testing.T) {
	// 4321, 1234
	divident, modular := 5, 3
	//a, m := modular, divident // Example: Find 3⁻¹ mod 7
	inverse, err := ModInverse(modular, divident)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Modular inverse of %d mod %d is %d\n", divident, modular, inverse)
	}
}

func TestMillerRabin(t *testing.T) {
	mainMiller()
}

func TestAes(t *testing.T) {
	mainAes()
}

func TestCRT(t *testing.T) {
	mainCRT()
}

func TestGmul(t *testing.T) {
	res := gmul(byte(5), byte(3))
	print(int(res))
}
func TestInverse(t *testing.T) {
	res := inverse(byte(8))
	print(int(res))
}

