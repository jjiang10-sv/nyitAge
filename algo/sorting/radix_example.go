package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const digit = 4
const maxbit = -1 << 31

// Constants are already defined in radix_sort.go

func mainRad() {
	// Simple example with mixed positive/negative numbers
	data := []int32{-2, 1, -1, 2}
	fmt.Printf("Original data: %v\n\n", data)

	radixSortWithTrace(data)

	fmt.Printf("\nFinal sorted result: %v\n", data)
}

func radixSortWithTrace(data []int32) {
	fmt.Println("=== STEP 1: Convert to bytes with sign bit flip ===")

	buf := bytes.NewBuffer(nil)
	ds := make([][]byte, len(data))

	for i, e := range data {
		fmt.Printf("Original: %d (binary: %032b)\n", e, uint32(e))
		fmt.Printf("%b", maxbit)
		//fmt.Sprintf("%b", e)
		flipped := e ^ maxbit
		fmt.Printf("After XOR:  %d (binary: %032b)\n", flipped, uint32(flipped))

		binary.Write(buf, binary.LittleEndian, flipped)
		b := make([]byte, digit)
		buf.Read(b)
		ds[i] = b

		fmt.Printf("As bytes: [%d, %d, %d, %d] (little-endian)\n", b[0], b[1], b[2], b[3])
		fmt.Println()
	}

	fmt.Println("=== STEP 2: Radix sort by each byte position ===")

	countingSort := make([][][]byte, 256)

	for bytePos := 0; bytePos < digit; bytePos++ {
		fmt.Printf("\n--- Processing byte position %d ---\n", bytePos)

		// Show current state
		fmt.Print("Current order: ")
		for i, b := range ds {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("[%d,%d,%d,%d]", b[0], b[1], b[2], b[3])
		}
		fmt.Println()

		// Distribute into buckets
		fmt.Printf("Sorting by byte[%d]:\n", bytePos)
		for i, b := range ds {
			bucketIndex := b[bytePos]
			countingSort[bucketIndex] = append(countingSort[bucketIndex], b)
			fmt.Printf("  Item %d: byte[%d]=%d -> bucket %d\n", i, bytePos, bucketIndex, bucketIndex)
		}

		// Show non-empty buckets
		fmt.Println("Non-empty buckets:")
		for k := 0; k < 256; k++ {
			if len(countingSort[k]) > 0 {
				fmt.Printf("  Bucket %d: ", k)
				for i, b := range countingSort[k] {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("[%d,%d,%d,%d]", b[0], b[1], b[2], b[3])
				}
				fmt.Println()
			}
		}

		// Collect back in order
		j := 0
		for k, bs := range countingSort {
			if len(bs) > 0 {
				copy(ds[j:], bs)
				j += len(bs)
				countingSort[k] = bs[:0]
			}
		}

		// Show result after this pass
		fmt.Print("After sorting: ")
		for i, b := range ds {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("[%d,%d,%d,%d]", b[0], b[1], b[2], b[3])
		}
		fmt.Println()
	}

	fmt.Println("\n=== STEP 3: Convert back to integers and restore sign ===")

	var w int32
	for i, b := range ds {
		buf.Write(b)
		binary.Read(buf, binary.LittleEndian, &w)
		restored := w ^ maxbit

		fmt.Printf("Bytes [%d,%d,%d,%d] -> %d -> %d (after XOR back)\n",
			b[0], b[1], b[2], b[3], w, restored)

		data[i] = restored
	}
}
