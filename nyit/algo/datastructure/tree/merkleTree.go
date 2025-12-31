package tree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Hash helper
func hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// NodeMerkle in Merkle Tree
type NodeMerkle struct {
	Hash  []byte
	Left  *NodeMerkle
	Right *NodeMerkle
}

func hash1(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

type NodeMerkle1 struct {
	Hash  []byte
	Left  *NodeMerkle1
	Right *NodeMerkle1
}

// MerkleTree struct
type MerkleTree struct {
	Root   *NodeMerkle
	Leaves []*NodeMerkle
}

type MerkleTree1 struct {
	Root   *NodeMerkle1
	Leaves []*NodeMerkle1
}

// NewMerkleTree builds a tree from raw leaf data
func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{}
	}

	leaves := make([]*NodeMerkle, len(data))
	for i, d := range data {
		leaves[i] = &NodeMerkle{Hash: hash(d)}
	}

	tree := &MerkleTree{Leaves: leaves}
	tree.Root = buildTree(leaves)
	return tree
}

func NewMerkleTree1(data [][]byte) *MerkleTree1 {
	if len(data) == 0 {
		return &MerkleTree1{}
	}
	leaves := make([]*NodeMerkle1, len(data))
	for i, d := range data {
		leaves[i] = &NodeMerkle1{Hash: hash(d)}
	}
	tree := &MerkleTree1{Leaves: leaves}
	tree.Root = buildTree1(leaves)
	return tree
}

// buildTree recursively creates parent levels
func buildTree(nodes []*NodeMerkle) *NodeMerkle {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var parents []*NodeMerkle
	for i := 0; i < len(nodes); i += 2 {
		if i+1 == len(nodes) {
			// odd node → duplicate last one
			parents = append(parents, merge(nodes[i], nodes[i]))
		} else {
			parents = append(parents, merge(nodes[i], nodes[i+1]))
		}
	}

	return buildTree(parents)
}
func buildTree1(nodes []*NodeMerkle1) *NodeMerkle1 {
	if len(nodes) == 1 {
		return nodes[0]
	}
	var parents []*NodeMerkle1
	for i := 0; i < len(nodes); i += 2 {

	}
}

// merge two child nodes
func merge(left, right *NodeMerkle) *NodeMerkle {
	combined := append(left.Hash, right.Hash...)
	return &NodeMerkle{
		Hash:  hash(combined),
		Left:  left,
		Right: right,
	}
}

// Append a new leaf & rebuild tree
func (t *MerkleTree) Append(data []byte) {
	leaf := &NodeMerkle{Hash: hash(data)}
	t.Leaves = append(t.Leaves, leaf)
	t.Root = buildTree(t.Leaves)
}

// List all leaf hashes (hex)
func (t *MerkleTree) ListLeaves() [][]byte {
	out := make([][]byte, len(t.Leaves))
	for i, leaf := range t.Leaves {
		out[i] = leaf.Hash
	}
	return out
}

///////////////////////////////////////////////////////////////////////////////
// MERKLE PROOF GENERATION
///////////////////////////////////////////////////////////////////////////////

// MerkleProof contains sibling hashes along the path to the root
type MerkleProof struct {
	Leaf     []byte
	Siblings [][]byte // each sibling hash
}

// GenerateProof builds a Merkle audit proof
func (t *MerkleTree) GenerateProof(index int) (*MerkleProof, error) {
	if index < 0 || index >= len(t.Leaves) {
		return nil, fmt.Errorf("invalid index")
	}

	var siblings [][]byte
	nodes := t.Leaves

	for len(nodes) > 1 {
		var nextLevel []*NodeMerkle

		for i := 0; i < len(nodes); i += 2 {
			var left, right *NodeMerkle
			left = nodes[i]
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = nodes[i] // duplicate last
			}

			// if the leaf is in this pair, save sibling
			if i == index || i+1 == index {
				if index == i {
					// sibling is right
					siblings = append(siblings, right.Hash)
				} else {
					siblings = append(siblings, left.Hash)
				}
				index = len(nextLevel) // move to parent index
			}

			nextLevel = append(nextLevel, merge(left, right))
		}

		nodes = nextLevel
	}

	return &MerkleProof{
		Leaf:     t.Leaves[index].Hash,
		Siblings: siblings,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// MERKLE PROOF VERIFICATION
///////////////////////////////////////////////////////////////////////////////

// VerifyProof validates a Merkle proof against a root hash
func VerifyProof(root []byte, proof *MerkleProof) bool {
	hashCur := proof.Leaf

	for _, s := range proof.Siblings {
		// try left-right combination
		leftRight := hash(append(hashCur, s...))
		rightLeft := hash(append(s, hashCur...))

		if bytes.Equal(leftRight, root) {
			return true
		}
		if bytes.Equal(rightLeft, root) {
			return true
		}

		// continue hashing upward (normally left + right order is known)
		hashCur = leftRight
	}
	return bytes.Equal(hashCur, root)
}
