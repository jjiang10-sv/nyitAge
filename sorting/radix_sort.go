package main

import (
	"bytes"
	"encoding/binary"
)

// const digit = 4
// const maxbit = -1 << 31

// func gotest() {
// 	var data = []int32{421, 15, -175, 90, -2, 214, -52, -166}
// 	fmt.Println("\n--- Unsorted --- \n\n", data)
// 	radixsort(data)
// 	fmt.Println("\n--- Sorted ---\n\n", data, "\n")
// }

// make the buckets,
// LSB :
// loop through
//
//	find it, put it in bucket
func radixsort(data []int32) {
	buf := bytes.NewBuffer(nil)
	ds := make([][]byte, len(data))
	for i, e := range data {
		binary.Write(buf, binary.LittleEndian, e^maxbit)
		b := make([]byte, digit)
		buf.Read(b)
		ds[i] = b
	}
	countingSort := make([][][]byte, 256)
	for i := 0; i < digit; i++ {
		for _, b := range ds {
			countingSort[b[i]] = append(countingSort[b[i]], b)
		}
		j := 0
		for k, bs := range countingSort {
			copy(ds[j:], bs)
			j += len(bs)
			countingSort[k] = bs[:0]
		}
	}
	var w int32
	for i, b := range ds {
		buf.Write(b)
		binary.Read(buf, binary.LittleEndian, &w)
		data[i] = w ^ maxbit
	}
}

// Radix sort is a non-comparative sorting algorithm that sorts data by processing individual digits or characters. Instead of comparing elements directly, it distributes elements into buckets based on each digit's value.

// How Radix Sort Works
// Basic Principle:

// Sort numbers digit by digit, starting from the least significant digit (LSD) or most significant digit (MSD)
// Use a stable sorting algorithm (like counting sort) as a subroutine for each digit position
// Repeat for each digit position until all digits are processed
// Algorithm Steps (LSD Radix Sort)
// Find the maximum number to determine the number of digits
// For each digit position (starting from rightmost):
// Use counting sort to sort elements based on current digit
// Maintain stability (preserve relative order of equal elements)
// Repeat until all digit positions are processed
// Example
// Sorting [170, 45, 75, 90, 2, 802, 24, 66]:

// Original: [170, 45, 75, 90, 2, 802, 24, 66]

// Sort by 1s place:
// [170, 90, 2, 802, 24, 45, 75, 66]

// Sort by 10s place:
// [2, 802, 24, 45, 66, 170, 75, 90]

// Sort by 100s place:
// [2, 24, 45, 66, 75, 90, 170, 802]
// Time & Space Complexity
// Time Complexity: O(d × (n + k))

// d = number of digits
// n = number of elements
// k = range of each digit (usually 10 for decimal)
// Space Complexity: O(n + k)

// Additional space for counting array and output array
// Advantages
// Linear time for fixed-width integers
// Stable sorting algorithm
// No comparisons needed
// Efficient for large datasets with small digit ranges
// Disadvantages
// Only works with integers or fixed-length strings
// Space overhead for counting arrays
// Performance depends on number of digits
// Not suitable for floating-point numbers without preprocessing
// Implementation Types
// LSD (Least Significant Digit):

// Processes digits from right to left
// More common and simpler to implement
// Good for fixed-width data
// MSD (Most Significant Digit):

// Processes digits from left to right
// Can stop early if prefixes differ
// Better for variable-length strings
// When to Use Radix Sort
// Good for:

// Large datasets of integers with limited digit count
// Fixed-width data (ZIP codes, phone numbers)
// When comparison-based sorts are too slow
// Avoid when:

// Data has variable lengths
// Working with floating-point numbers
// Small datasets (overhead not worth it)
// Memory is constrained
// Radix sort shines when you have lots of integers with a predictable number of digits - it can beat O(n log n) comparison sorts by avoiding comparisons entirely.
