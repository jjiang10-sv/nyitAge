package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// GenerateSalt generates a random salt of a given length
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// HashWithSalt computes the SHA-512 hash of an input with a salt
func HashWithSalt(input string, salt []byte) string {
	// Append the salt to the input
	saltedInput := append([]byte(input), salt...)

	// Compute the SHA-512 hash of the salted input
	hasher := sha512.New()
	hasher.Write(saltedInput)
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}

func main() {
	// Example input
	input := "Hello, World!"

	// Generate a random salt
	salt, err := GenerateSalt(16)
	if err != nil {
		fmt.Printf("Error generating salt: %v\n", err)
		return
	}

	// Compute the salted hash
	hash := HashWithSalt(input, salt)

	// Print the results
	fmt.Printf("Input: %s\n", input)
	fmt.Printf("Salt: %x\n", salt)
	fmt.Printf("Salted SHA-512 Hash: %s\n", hash)
}

// HashFileSHA512 computes the SHA-512 hash of a file
func HashFileSHA512(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA-512 hasher
	hasher := sha512.New()

	// Copy the file content into the hasher
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	// Compute the hash and convert it to a hexadecimal string
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

func mainHashFile() {
	// Example file path
	filePath := "example.txt"

	// Compute the SHA-512 hash of the file
	hash, err := HashFileSHA512(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the hash
	fmt.Printf("SHA-512 Hash of %s: %s\n", filePath, hash)
}

// SimulateHashFunction simulates a hash function for demonstration purposes
func SimulateHashFunction(key string, numBuckets int) int {
	// Use SHA-256 to compute a hash value for the key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)

	// Convert the hash to an integer
	hashInt := int(hashBytes[0])<<24 | int(hashBytes[1])<<16 | int(hashBytes[2])<<8 | int(hashBytes[3])

	// Use modulo to determine the bucket index
	bucketIndex := hashInt % numBuckets
	return bucketIndex
}

func hashFuncClustemain() {
	// Simulate a hash table with 10 buckets
	numBuckets := 10
	buckets := make([][]string, numBuckets)

	// Keys to insert into the hash table
	keys := []string{
		"apple", "apricot", "banana", "blueberry", "cherry",
		"date", "elderberry", "fig", "grape", "honeydew",
		"kiwi", "lemon", "lime", "mango", "nectarine",
		"orange", "papaya", "peach", "pear", "pineapple",
	}

	// Insert keys into the hash table
	for _, key := range keys {
		bucketIndex := SimulateHashFunction(key, numBuckets)
		buckets[bucketIndex] = append(buckets[bucketIndex], key)
	}

	// Print the distribution of keys across buckets
	for i, bucket := range buckets {
		fmt.Printf("Bucket %d: %v\n", i, bucket)
	}

	// Calculate and print the clustering factor
	clusteringFactor := calculateClusteringFactor(buckets)
	fmt.Printf("\nClustering Factor: %.2f\n", clusteringFactor)
}

// calculateClusteringFactor calculates the clustering factor of the hash table
func calculateClusteringFactor(buckets [][]string) float64 {
	totalKeys := 0
	occupiedBuckets := 0
	maxChainLength := 0

	for _, bucket := range buckets {
		if len(bucket) > 0 {
			occupiedBuckets++
			totalKeys += len(bucket)
			if len(bucket) > maxChainLength {
				maxChainLength = len(bucket)
			}
		}
	}

	// Clustering factor = (Average chain length) / (Load factor)
	averageChainLength := float64(totalKeys) / float64(occupiedBuckets)
	loadFactor := float64(totalKeys) / float64(len(buckets))
	clusteringFactor := averageChainLength / loadFactor

	return clusteringFactor
}

const (
	numBuckets = 10 // Size of the hash table
)

// HashTable represents a hash table with double hashing
type HashTable struct {
	buckets [numBuckets]string
}

// Hash1 computes the primary hash value for a key
func Hash1(key string) int {
	// Use SHA-256 to compute a hash value for the key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)

	// Convert the hash to an integer
	hashInt := int(hashBytes[0])<<24 | int(hashBytes[1])<<16 | int(hashBytes[2])<<8 | int(hashBytes[3])

	// Use modulo to determine the bucket index
	bucketIndex := hashInt % numBuckets
	return bucketIndex
}

// Hash2 computes the secondary hash value for a key
func Hash2(key string) int {
	// Use a different hash function (e.g., SHA-1 or a simple custom hash)
	// Here, we use a simple custom hash for demonstration purposes
	hash := 0
	for _, char := range key {
		hash += int(char)
	}
	// Ensure the secondary hash is non-zero and relatively prime to numBuckets
	return hash%7 + 1 // Example: Use a prime number less than numBuckets
}

// Insert inserts a key into the hash table using double hashing
func (ht *HashTable) Insert(key string) bool {
	initialIndex := Hash1(key)
	stepSize := Hash2(key)

	for i := 0; i < numBuckets; i++ {
		// Compute the index using double hashing
		index := (initialIndex + i*stepSize) % numBuckets

		// If the bucket is empty, insert the key
		if ht.buckets[index] == "" {
			ht.buckets[index] = key
			fmt.Printf("Inserted '%s' at bucket %d\n", key, index)
			return true
		}

		// If the key already exists, do not insert
		if ht.buckets[index] == key {
			fmt.Printf("Key '%s' already exists at bucket %d\n", key, index)
			return false
		}
	}

	// If no empty bucket is found after probing, the hash table is full
	fmt.Printf("Could not insert '%s': Hash table is full\n", key)
	return false
}

// Search searches for a key in the hash table using double hashing
func (ht *HashTable) Search(key string) (int, bool) {
	initialIndex := Hash1(key)
	stepSize := Hash2(key)

	for i := 0; i < numBuckets; i++ {
		// Compute the index using double hashing
		index := (initialIndex + i*stepSize) % numBuckets

		// If the bucket contains the key, return the index
		if ht.buckets[index] == key {
			return index, true
		}

		// If an empty bucket is found, the key does not exist
		if ht.buckets[index] == "" {
			return -1, false
		}
	}

	// If the key is not found after probing, it does not exist
	return -1, false
}

func doubleHashMain() {
	// Create a hash table
	ht := HashTable{}

	// Insert keys into the hash table
	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "kiwi", "lemon"}
	for _, key := range keys {
		ht.Insert(key)
	}

	// Search for keys in the hash table
	searchKeys := []string{"apple", "banana", "mango", "kiwi"}
	for _, key := range searchKeys {
		index, found := ht.Search(key)
		if found {
			fmt.Printf("Found '%s' at bucket %d\n", key, index)
		} else {
			fmt.Printf("Key '%s' not found\n", key)
		}
	}

	// Print the final state of the hash table
	fmt.Println("\nFinal Hash Table:")
	for i, bucket := range ht.buckets {
		if bucket != "" {
			fmt.Printf("Bucket %d: %s\n", i, bucket)
		}
	}
}

// Node represents a node in the linked list
type Node struct {
	key  string
	next *Node
}

// HashTableChain represents a hash table with separate chaining
type HashTableChain struct {
	buckets [numBuckets]*Node
}

// HashFunction computes the hash value for a key
func HashFunction(key string) int {
	// Use SHA-256 to compute a hash value for the key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)

	// Convert the hash to an integer
	hashInt := int(hashBytes[0])<<24 | int(hashBytes[1])<<16 | int(hashBytes[2])<<8 | int(hashBytes[3])

	// Use modulo to determine the bucket index
	bucketIndex := hashInt % numBuckets
	return bucketIndex
}

// Insert inserts a key into the hash table using separate chaining
func (ht *HashTableChain) Insert(key string) {
	index := HashFunction(key)

	// Create a new node
	newNode := &Node{key: key, next: nil}

	// If the bucket is empty, insert the node
	if ht.buckets[index] == nil {
		ht.buckets[index] = newNode
	} else {
		// Traverse the linked list and append the new node
		current := ht.buckets[index]
		for current.next != nil {
			// If the key already exists, do not insert
			if current.key == key {
				fmt.Printf("Key '%s' already exists at bucket %d\n", key, index)
				return
			}
			current = current.next
		}
		// Check the last node
		if current.key == key {
			fmt.Printf("Key '%s' already exists at bucket %d\n", key, index)
			return
		}
		// Append the new node
		current.next = newNode
	}

	fmt.Printf("Inserted '%s' at bucket %d\n", key, index)
}

// Search searches for a key in the hash table using separate chaining
func (ht *HashTableChain) Search(key string) (int, bool) {
	index := HashFunction(key)

	// Traverse the linked list at the bucket
	current := ht.buckets[index]
	for current != nil {
		if current.key == key {
			return index, true
		}
		current = current.next
	}

	// If the key is not found, return false
	return -1, false
}

// PrintHashTable prints the contents of the hash table
func (ht *HashTableChain) PrintHashTable() {
	for i, bucket := range ht.buckets {
		if bucket != nil {
			fmt.Printf("Bucket %d: ", i)
			current := bucket
			for current != nil {
				fmt.Printf("%s -> ", current.key)
				current = current.next
			}
			fmt.Println("nil")
		}
	}
}

func chinaHashtableMain() {
	// Create a hash table
	ht := HashTableChain{}

	// Insert keys into the hash table
	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "kiwi", "lemon"}
	for _, key := range keys {
		ht.Insert(key)
	}

	// Search for keys in the hash table
	searchKeys := []string{"apple", "banana", "mango", "kiwi"}
	for _, key := range searchKeys {
		index, found := ht.Search(key)
		if found {
			fmt.Printf("Found '%s' at bucket %d\n", key, index)
		} else {
			fmt.Printf("Key '%s' not found\n", key)
		}
	}

	// Print the final state of the hash table
	fmt.Println("\nFinal Hash Table:")
	ht.PrintHashTable()
}

const (
	loadFactorThreshold = 0.7 // Threshold for choosing separate chaining
)

// SimulateOpenAddressing simulates open addressing (linear probing)
func SimulateOpenAddressing(keys []string) {
	fmt.Println("Using Open Addressing (Linear Probing)...")
	buckets := make([]string, numBuckets)

	for _, key := range keys {
		index := HashFunction(key)

		// Linear probing to find an empty bucket
		for i := 0; i < numBuckets; i++ {
			probeIndex := (index + i) % numBuckets
			if buckets[probeIndex] == "" {
				buckets[probeIndex] = key
				fmt.Printf("Inserted '%s' at bucket %d\n", key, probeIndex)
				break
			}
		}
	}

	// Print the final state of the hash table
	fmt.Println("\nFinal Hash Table (Open Addressing):")
	for i, bucket := range buckets {
		if bucket != "" {
			fmt.Printf("Bucket %d: %s\n", i, bucket)
		}
	}
}

// SimulateSeparateChaining simulates separate chaining
func SimulateSeparateChaining(keys []string) {
	fmt.Println("Using Separate Chaining...")
	buckets := make([][]string, numBuckets)

	for _, key := range keys {
		index := HashFunction(key)

		// Append the key to the linked list in the bucket
		buckets[index] = append(buckets[index], key)
		fmt.Printf("Inserted '%s' at bucket %d\n", key, index)
	}

	// Print the final state of the hash table
	fmt.Println("\nFinal Hash Table (Separate Chaining):")
	for i, bucket := range buckets {
		if len(bucket) > 0 {
			fmt.Printf("Bucket %d: %v\n", i, bucket)
		}
	}
}

// DecideCollisionResolution decides whether to use open addressing or separate chaining
func DecideCollisionResolution(keys []string) {
	// Calculate the load factor
	loadFactor := float64(len(keys)) / float64(numBuckets)

	// Decide based on the load factor
	if loadFactor < loadFactorThreshold {
		fmt.Printf("Load factor (%.2f) is low. Choosing Open Addressing.\n", loadFactor)
		SimulateOpenAddressing(keys)
	} else {
		fmt.Printf("Load factor (%.2f) is high. Choosing Separate Chaining.\n", loadFactor)
		SimulateSeparateChaining(keys)
	}
}

func decisionFactorMain() {
	// Example keys to insert into the hash table
	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "kiwi", "lemon"}

	// Decide and simulate the collision resolution strategy
	DecideCollisionResolution(keys)
}
