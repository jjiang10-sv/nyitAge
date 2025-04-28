package main

import (
	"fmt"
)

// Recursive function to solve Tower of Hanoi
func hanoi(n int, source, auxiliary, target string) {
	if n > 0 {
		// Move n-1 disks from source to auxiliary using target as helper
		hanoi(n-1, source, target, auxiliary)

		// Move the nth disk from source to target
		fmt.Printf("Move disk %d from %s to %s\n", n, source, target)

		// Move n-1 disks from auxiliary to target using source as helper
		hanoi(n-1, auxiliary, source, target)
	}
}
