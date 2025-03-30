// Here's a Go implementation of a blockchain that handles stale chains and orphan blocks:

// ```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

type Blockchain struct {
	mainChain  []*Block
	orphans    map[string]*Block // Orphan blocks by their hash
	difficulty int
}

func NewBlockchain(difficulty int) *Blockchain {
	return &Blockchain{
		mainChain:  make([]*Block, 0),
		orphans:    make(map[string]*Block),
		difficulty: difficulty,
	}
}

func (bc *Blockchain) AddBlock(newBlock *Block) error {
	if !bc.ValidateBlock(newBlock) {
		return errors.New("invalid block")
	}

	// Handle genesis block
	if len(bc.mainChain) == 0 {
		if newBlock.Index == 0 && newBlock.PrevHash == "" {
			bc.mainChain = append(bc.mainChain, newBlock)
			return nil
		}
		return errors.New("invalid genesis block")
	}

	currentTip := bc.mainChain[len(bc.mainChain)-1]

	// Try to add to main chain
	if newBlock.PrevHash == currentTip.Hash {
		bc.mainChain = append(bc.mainChain, newBlock)
		bc.checkOrphans()
		return nil
	}

	// Check if it creates a fork
	for i, block := range bc.mainChain {
		if block.Hash == newBlock.PrevHash {
			forkChain := append([]*Block{}, bc.mainChain[:i+1]...)
			forkChain = append(forkChain, newBlock)

			// Extend the fork with orphans if possible
			bc.extendChain(forkChain)

			if bc.ValidateChain(forkChain) && len(forkChain) > len(bc.mainChain) {
				bc.mainChain = forkChain
				bc.checkOrphans()
				return nil
			}
		}
	}

	// Check if it extends an orphan
	if parent, exists := bc.orphans[newBlock.PrevHash]; exists {
		chain := []*Block{parent, newBlock}
		bc.extendChain(chain)

		if bc.ValidateChain(chain) {
			// Check if this orphan chain is longer than main
			if commonAncestor, depth := bc.findCommonAncestor(chain); commonAncestor != -1 {
				totalLength := depth + len(chain)
				if totalLength > len(bc.mainChain) {
					newMain := append(bc.mainChain[:commonAncestor+1], chain...)
					bc.mainChain = newMain

					// Remove used orphans
					for _, b := range chain {
						delete(bc.orphans, b.Hash)
					}
					bc.checkOrphans()
					return nil
				}
			}
		}
	}

	// Add to orphans
	bc.orphans[newBlock.Hash] = newBlock
	return nil
}

func (bc *Blockchain) extendChain(chain []*Block) {
	for {
		lastHash := chain[len(chain)-1].Hash
		if nextBlock, exists := bc.orphans[lastHash]; exists {
			chain = append(chain, nextBlock)
			delete(bc.orphans, lastHash)
		} else {
			break
		}
	}
}

func (bc *Blockchain) findCommonAncestor(chain []*Block) (int, int) {
	for i := len(bc.mainChain) - 1; i >= 0; i-- {
		for j, block := range chain {
			if bc.mainChain[i].Hash == block.PrevHash {
				return i, j + 1 // Return ancestor index and depth
			}
		}
	}
	return -1, 0
}

func (bc *Blockchain) checkOrphans() {
	for hash, block := range bc.orphans {
		if block.PrevHash == bc.mainChain[len(bc.mainChain)-1].Hash {
			delete(bc.orphans, hash)
			_ = bc.AddBlock(block)
		}
	}
}

func (bc *Blockchain) ValidateBlock(block *Block) bool {
	if block.Hash != calculateHash(block) {
		return false
	}
	if !strings.HasPrefix(block.Hash, strings.Repeat("0", bc.difficulty)) {
		return false
	}
	return true
}

func (bc *Blockchain) ValidateChain(chain []*Block) bool {
	if len(chain) == 0 {
		return false
	}

	// Check genesis block
	if chain[0].Index != 0 || chain[0].PrevHash != "" {
		return false
	}

	for i := 1; i < len(chain); i++ {
		if chain[i].PrevHash != chain[i-1].Hash {
			return false
		}
		if !bc.ValidateBlock(chain[i]) {
			return false
		}
	}
	return true
}

func calculateHash(block *Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Data + block.PrevHash + strconv.Itoa(block.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func mineBlock(prevBlock *Block, data string, difficulty int) *Block {
	nonce := 0
	for {
		block := &Block{
			Index:     prevBlock.Index + 1,
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      data,
			PrevHash:  prevBlock.Hash,
			Nonce:     nonce,
		}
		block.Hash = calculateHash(block)
		if strings.HasPrefix(block.Hash, strings.Repeat("0", difficulty)) {
			return block
		}
		nonce++
	}
}

func staleOrphanMain() {
	// Create blockchain with difficulty of 2
	bc := NewBlockchain(2)

	// Create and add genesis block
	genesis := &Block{
		Index:     0,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      "Genesis Block",
		PrevHash:  "",
		Nonce:     0,
	}
	genesis.Hash = calculateHash(genesis)
	bc.AddBlock(genesis)

	// Mine and add blocks
	block1 := mineBlock(bc.mainChain[0], "Block 1 Data", bc.difficulty)
	bc.AddBlock(block1)

	block2 := mineBlock(block1, "Block 2 Data", bc.difficulty)
	bc.AddBlock(block2)

	// Create a fork
	block3a := mineBlock(block2, "Block 3A Data", bc.difficulty)
	block3b := mineBlock(block2, "Block 3B Data", bc.difficulty)

	bc.AddBlock(block3a)
	fmt.Println("Main chain after adding 3A:")
	for _, block := range bc.mainChain {
		fmt.Printf("Index: %d, Hash: %s\n", block.Index, block.Hash)
	}

	bc.AddBlock(block3b)
	fmt.Println("\nOrphans after adding 3B:", len(bc.orphans))

	// Extend the orphan chain
	block4b := mineBlock(block3b, "Block 4B Data", bc.difficulty)
	bc.AddBlock(block4b)

	fmt.Println("\nMain chain after orphan chain becomes longer:")
	for _, block := range bc.mainChain {
		fmt.Printf("Index: %d, Hash: %s\n", block.Index, block.Hash)
	}
	fmt.Println("Remaining orphans:", len(bc.orphans))
}

// ```

// This implementation demonstrates:

// 1. **Blockchain Structure**: Maintains a main chain and orphan blocks
// 2. **Proof-of-Work**: Mining with adjustable difficulty
// 3. **Orphan Blocks**: Handling of blocks that don't immediately connect to the main chain
// 4. **Chain Reorganization**: Automatically switches to longer valid chains
// 5. **Stale Chains**: Previous main chain becomes stale when a longer chain is found

// Key features:
// - Orphan blocks are stored separately and rechecked when the chain changes
// - Automatic chain reorganization when a longer valid chain is found
// - Proof-of-work validation
// - Proper block linking and hash validation

// When you run this code, you'll see:
// 1. Initial main chain construction
// 2. Orphan block creation when a fork occurs
// 3. Automatic switch to the longer chain when it becomes available
// 4. Orphan blocks being cleared when they become part of the main chain

// The code shows how a blockchain can handle forks and orphan blocks while maintaining consensus through the longest chain rule.
