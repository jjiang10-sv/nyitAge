package main

import (
	"fmt"
)

// Function to perform binary division using bitwise operations
func binaryDivision(dividend, generator int) (int, int) {
	// Determine the length of the generator
	genLen := 0
	for temp := generator; temp > 0; temp >>= 1 {
		genLen++
	}

	// Append (genLen - 1) zeros to the dividend
	dividend <<= (genLen - 1)

	// Perform division
	for i := 31; i >= genLen-1; i-- {
		if (dividend>>i)&1 == 1 {
			// XOR operation for each bit
			dividend ^= generator << (i - genLen + 1)
		}
	}

	// The remainder is the last (genLen - 1) bits of the dividend
	remainder := dividend & ((1 << (genLen - 1)) - 1)
	return remainder, dividend
}

func mainbit() {
	// Original bit stream and generator (in decimal form)
	bitStream := 0b10011101 // Binary: 10011101
	generator := 0b1001     // Binary: 1001

	// Sender calculates CRC
	remainder, _ := binaryDivision(bitStream, generator)
	transmitted := (bitStream << (3)) | remainder // Append CRC to the bit stream

	// Introduce error: Invert the third bit from the left
	received := transmitted ^ (1 << 8) // Flip the third bit (from the left)

	// Receiver checks for errors
	receiverRemainder, _ := binaryDivision(received, generator)

	// Output the results
	fmt.Printf("Original Bit Stream: %b\n", bitStream)
	fmt.Printf("Transmitted Bit Stream: %b\n", transmitted)
	fmt.Printf("Received Bit Stream (with error): %b\n", received)
	fmt.Printf("Receiver Remainder: %b\n", receiverRemainder)

	// Check for errors
	if receiverRemainder != 0 {
		fmt.Println("Error detected in transmission!")
	} else {
		fmt.Println("No error detected.")
	}
}
