package main

import "fmt"

func mainByte() {
	// Example 1: Simple []byte slice
	fmt.Println("=== Example 1: []byte basics ===")

	// Create a []byte slice
	b1 := []byte{65, 66, 67, 68} // ASCII for A, B, C, D
	fmt.Printf("b1: %v\n", b1)
	fmt.Printf("b1 as string: %s\n", string(b1))
	fmt.Printf("b1[0]: %d (which is '%c')\n", b1[0], b1[0])

	// Another []byte
	b2 := []byte("Hello")
	fmt.Printf("b2: %v\n", b2)
	fmt.Printf("b2 as string: %s\n", string(b2))

	fmt.Println()

	// Example 2: [][]byte (slice of byte slices)
	fmt.Println("=== Example 2: [][]byte (2D) ===")

	words := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
	}

	fmt.Printf("words: %v\n", words)
	for i, word := range words {
		fmt.Printf("words[%d]: %v (string: %s)\n", i, word, string(word))
	}

	fmt.Println()

	// Example 3: [][][]byte (like countingSort in radix sort)
	fmt.Println("=== Example 3: [][][]byte (3D - like countingSort) ===")

	// Create a simplified version with just 4 buckets instead of 256
	buckets := make([][][]byte, 4)

	// Add some byte slices to different buckets
	buckets[0] = append(buckets[0], []byte{1, 2, 3, 4})
	buckets[0] = append(buckets[0], []byte{5, 6, 7, 8})
	buckets[2] = append(buckets[2], []byte{9, 10, 11, 12})

	fmt.Printf("buckets structure: %v\n", buckets)

	// Show each bucket
	for i, bucket := range buckets {
		if len(bucket) > 0 {
			fmt.Printf("Bucket %d has %d items:\n", i, len(bucket))
			for j, item := range bucket {
				fmt.Printf("  Item %d: %v\n", j, item)
			}
		} else {
			fmt.Printf("Bucket %d is empty\n", i)
		}
	}

	fmt.Println()

	// Example 4: How it's used in radix sort
	fmt.Println("=== Example 4: Radix sort usage simulation ===")

	// Simulate the radix sort data structure
	countingSort := make([][][]byte, 256) // 256 buckets for byte values 0-255

	// Some example 4-byte arrays (like in radix sort)
	data := [][]byte{
		{254, 255, 255, 127}, // represents -2 after XOR
		{1, 0, 0, 128},       // represents 1 after XOR
		{255, 255, 255, 127}, // represents -1 after XOR
		{2, 0, 0, 128},       // represents 2 after XOR
	}

	// Simulate sorting by first byte (byte position 0)
	fmt.Println("Distributing by first byte:")
	for i, item := range data {
		bucketIndex := item[0] // Use first byte as bucket index
		countingSort[bucketIndex] = append(countingSort[bucketIndex], item)
		fmt.Printf("Item %d: %v -> bucket %d\n", i, item, bucketIndex)
	}

	// Show non-empty buckets
	fmt.Println("\nNon-empty buckets:")
	for k := 0; k < 256; k++ {
		if len(countingSort[k]) > 0 {
			fmt.Printf("Bucket %d: %v\n", k, countingSort[k])
		}
	}
}
