package tree

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMerkleTreeBasic(t *testing.T) {
	data := [][]byte{
		[]byte("A"),
		[]byte("B"),
		[]byte("C"),
		[]byte("D"),
	}

	tree := NewMerkleTree(data)

	if tree.Root == nil {
		t.Fatalf("expected non-nil root")
	}

	if len(tree.Leaves) != 4 {
		t.Fatalf("expected 4 leaves, got %d", len(tree.Leaves))
	}

	// print root for debugging
	t.Logf("Root: %s", hex.EncodeToString(tree.Root.Hash))
}

func TestMerkleTreeAppend(t *testing.T) {
	data := [][]byte{
		[]byte("A"),
		[]byte("B"),
	}

	tree := NewMerkleTree(data)
	origRoot := tree.Root.Hash

	tree.Append([]byte("C"))

	if bytes.Equal(origRoot, tree.Root.Hash) {
		t.Fatalf("expected different root after append")
	}

	if len(tree.Leaves) != 3 {
		t.Fatalf("expected 3 leaves, got %d", len(tree.Leaves))
	}

	t.Logf("New root: %s", hex.EncodeToString(tree.Root.Hash))
}

func TestMerkleProofValid(t *testing.T) {
	data := [][]byte{
		[]byte("A"),
		[]byte("B"),
		[]byte("C"),
		[]byte("D"),
	}

	tree := NewMerkleTree(data)

	for i := range data {
		proof, err := tree.GenerateProof(i)
		if err != nil {
			t.Fatalf("proof error: %v", err)
		}

		if !VerifyProof(tree.Root.Hash, proof) {
			t.Fatalf("valid proof failed for index %d", i)
		}
	}
}

func TestMerkleProofInvalid(t *testing.T) {
	data := [][]byte{
		[]byte("A"),
		[]byte("B"),
		[]byte("C"),
	}

	tree := NewMerkleTree(data)

	proof, err := tree.GenerateProof(1)
	if err != nil {
		t.Fatalf("proof error: %v", err)
	}

	// Tamper leaf
	proof.Leaf = []byte("tampered")

	if VerifyProof(tree.Root.Hash, proof) {
		t.Fatalf("proof verification should fail for tampered proof")
	}
}

func TestOddNumberLeaves(t *testing.T) {
	data := [][]byte{
		[]byte("A"),
		[]byte("B"),
		[]byte("C"),
	}

	tree := NewMerkleTree(data)
	rootOdd := tree.Root.Hash

	// Now compute manually the expected tree:

	// Leaves
	hA := hash([]byte("A"))
	hB := hash([]byte("B"))
	hC := hash([]byte("C"))

	// Level 1 (pairs): (A,B) and (C,C)
	ab := hash(append(hA, hB...))
	cc := hash(append(hC, hC...))

	// Level 2 (root)
	expectedRoot := hash(append(ab, cc...))

	if !bytes.Equal(rootOdd, expectedRoot) {
		t.Fatalf("expected root %x, got %x", expectedRoot, rootOdd)
	}
}

func TestGenerateProofIndexBounds(t *testing.T) {
	tree := NewMerkleTree([][]byte{
		[]byte("A"),
	})

	// negative index
	if _, err := tree.GenerateProof(-1); err == nil {
		t.Fatal("expected error for negative index")
	}

	// too large index
	if _, err := tree.GenerateProof(2); err == nil {
		t.Fatal("expected error for invalid index")
	}
}

func TestTraverseLeaves(t *testing.T) {
	data := [][]byte{
		[]byte("X"),
		[]byte("Y"),
		[]byte("Z"),
	}

	tree := NewMerkleTree(data)
	leaves := tree.ListLeaves()

	if len(leaves) != len(data) {
		t.Fatalf("expected %d leaves, got %d", len(data), len(leaves))
	}

	for i, leaf := range leaves {
		if !bytes.Equal(leaf, hash(data[i])) {
			t.Fatalf("leaf mismatch at %d", i)
		}
	}
}
