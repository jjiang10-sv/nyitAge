package main

import (
	"fmt"
)

func calculateSum(numbers []int, result *int) {
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	*result = sum
}

func main11() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	mid := len(numbers) / 2
	//wg := sync.WaitGroup{}
	//chan :=
	var finalSum int
	go calculateSum(numbers[:mid], &finalSum)
	go calculateSum(numbers[mid:], &finalSum)

	fmt.Printf("Total Sum: %d\n", finalSum)
}
