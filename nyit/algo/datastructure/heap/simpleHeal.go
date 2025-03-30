package main

import (
	"fmt"
)

// MinHeap struct
type MinHeap struct {
	arr []int
}

// Insert a value into the heap
func (h *MinHeap) Insert(val int) {
	h.arr = append(h.arr, val)
	h.heapifyUp(len(h.arr) - 1)
}

// ExtractMin removes and returns the smallest element
func (h *MinHeap) ExtractMin() int {
	if len(h.arr) == 0 {
		fmt.Println("Heap is empty!")
		return -1
	}

	min := h.arr[0]
	h.arr[0] = h.arr[len(h.arr)-1] // Move last element to root
	h.arr = h.arr[:len(h.arr)-1]   // Remove last element
	h.heapifyDown(0)

	return min
}

// Heapify Up (used during Insert)
func (h *MinHeap) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if h.arr[parent] > h.arr[index] {
			h.arr[parent], h.arr[index] = h.arr[index], h.arr[parent] // Swap
			index = parent
		} else {
			break
		}
	}
}

// Heapify Down (used during ExtractMin)
func (h *MinHeap) heapifyDown(index int) {
	lastIndex := len(h.arr) - 1
	for {
		left := 2*index + 1
		right := 2*index + 2
		smallest := index

		if left <= lastIndex && h.arr[left] < h.arr[smallest] {
			smallest = left
		}
		if right <= lastIndex && h.arr[right] < h.arr[smallest] {
			smallest = right
		}
		if smallest != index {
			h.arr[index], h.arr[smallest] = h.arr[smallest], h.arr[index] // Swap
			index = smallest
		} else {
			break
		}
	}
}

// Print heap
func (h *MinHeap) PrintHeap() {
	fmt.Println(h.arr)
}

// Main function to test the heap
func main() {
	h := &MinHeap{}

	h.Insert(10)
	h.Insert(20)
	h.Insert(5)
	h.Insert(6)
	h.Insert(1)

	fmt.Println("Heap after inserts:")
	h.PrintHeap()

	fmt.Println("Extracted Min:", h.ExtractMin())
	fmt.Println("Heap after extraction:")
	h.PrintHeap()
}
