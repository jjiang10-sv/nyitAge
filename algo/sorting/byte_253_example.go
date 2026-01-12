package main

import (
	"encoding/binary"
	"fmt"
)

func mainBytesd() {
	fmt.Println("=== Different ways to represent 253 with []byte ===")

	// Method 1: Single byte (most common for 253)
	fmt.Println("1. Single byte representation:")
	b1 := []byte{253}
	fmt.Printf("   []byte{253} = %v\n", b1)
	fmt.Printf("   As integer: %d\n", b1[0])
	fmt.Printf("   As binary: %08b\n", b1[0])
	fmt.Println()

	// Method 2: String representation (if 253 were ASCII)
	fmt.Println("2. String representation:")
	b2 := []byte("253")
	fmt.Printf("   []byte(\"253\") = %v\n", b2)
	fmt.Printf("   As string: %s\n", string(b2))
	fmt.Printf("   Individual bytes: %d, %d, %d (ASCII for '2', '5', '3')\n", b2[0], b2[1], b2[2])
	fmt.Println()

	// Method 3: Multi-byte integer representations
	fmt.Println("3. Multi-byte integer representations:")

	// 16-bit (2 bytes)
	b3 := make([]byte, 2)
	binary.LittleEndian.PutUint16(b3, 253)
	fmt.Printf("   16-bit little-endian: %v\n", b3)

	b4 := make([]byte, 2)
	binary.BigEndian.PutUint16(b4, 253)
	fmt.Printf("   16-bit big-endian: %v\n", b4)

	// 32-bit (4 bytes)
	b5 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b5, 253)
	fmt.Printf("   32-bit little-endian: %v\n", b5)

	b6 := make([]byte, 4)
	binary.BigEndian.PutUint32(b6, 253)
	fmt.Printf("   32-bit big-endian: %v\n", b6)

	// 64-bit (8 bytes)
	b7 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b7, 253)
	fmt.Printf("   64-bit little-endian: %v\n", b7)
	fmt.Println()

	// Method 4: Manual byte construction
	fmt.Println("4. Manual construction:")
	// 253 in binary is 11111101
	// This fits in one byte, but we can pad with zeros
	b8 := []byte{0, 0, 0, 253} // 4 bytes with leading zeros
	fmt.Printf("   []byte{0, 0, 0, 253} = %v\n", b8)
	fmt.Println()

	// Method 5: Verification - convert back to integers
	fmt.Println("5. Converting back to integers:")
	fmt.Printf("   Single byte %v -> %d\n", b1, b1[0])
	fmt.Printf("   16-bit LE %v -> %d\n", b3, binary.LittleEndian.Uint16(b3))
	fmt.Printf("   16-bit BE %v -> %d\n", b4, binary.BigEndian.Uint16(b4))
	fmt.Printf("   32-bit LE %v -> %d\n", b5, binary.LittleEndian.Uint32(b5))
	fmt.Printf("   32-bit BE %v -> %d\n", b6, binary.BigEndian.Uint32(b6))
	fmt.Println()

	// Method 6: In context of radix sort
	fmt.Println("6. In radix sort context (like your example):")
	// If 253 appeared as a byte value in the radix sort
	fmt.Printf("   If byte value is 253, it goes to bucket 253\n")
	fmt.Printf("   countingSort[253] would contain items with byte value 253\n")

	// Example: a 4-byte array where one byte is 253
	example := []byte{253, 100, 50, 200}
	fmt.Printf("   Example 4-byte array: %v\n", example)
	fmt.Printf("   First byte (253) determines bucket: %d\n", example[0])
}
