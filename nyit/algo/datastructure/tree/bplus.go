package tree

import (
	"fmt"
	"sort"
)

type Record struct {
	Key   int
	Value interface{}
}

type Block struct {
	Entries  []Record
	Overflow *Block
}

type ISAM struct {
	Index     []int // Sorted max keys of each primary block
	Primary   []*Block
	BlockSize int
}

func NewISAM(records []Record, blockSize int) *ISAM {
	sort.Slice(records, func(i, j int) bool { return records[i].Key < records[j].Key })

	var index []int
	var primary []*Block
	// sort the records based on the key
	// divide the records by blocksize into blocks and append the block into primary
	// add the key of last record in the block into the index array
	for i := 0; i < len(records); i += blockSize {
		end := i + blockSize
		if end > len(records) {
			end = len(records)
		}
		block := &Block{Entries: records[i:end]}
		primary = append(primary, block)
		if len(block.Entries) > 0 {
			index = append(index, block.Entries[len(block.Entries)-1].Key)
		}
	}
	return &ISAM{Index: index, Primary: primary, BlockSize: blockSize}
}

// find the index of the key in the index array
// if less than isam index length, then return the block in primary
// if equal, then return last block in the primary
func (isam *ISAM) FindBlock(key int) *Block {
	i := sort.Search(len(isam.Index), func(i int) bool { return isam.Index[i] >= key })
	if i < len(isam.Index) {
		return isam.Primary[i]
	}
	return isam.Primary[len(isam.Primary)-1]
}

func (isam *ISAM) Insert(key int, value interface{}) {
	// find the block
	block := isam.FindBlock(key)
	if len(block.Entries) < isam.BlockSize {
		// if the block has space for more record, then find the place to insert
		// all sort out by the key, so it is easy to find the place
		i := sort.Search(len(block.Entries), func(i int) bool { return block.Entries[i].Key >= key })
		block.Entries = append(block.Entries[:i], append([]Record{{key, value}}, block.Entries[i:]...)...)
	} else {
		current := block
		// if the block has no record space, then deal with overflow
		for {
			// if no overflow block,, then create one with the k,v
			if current.Overflow == nil {
				current.Overflow = &Block{Entries: []Record{{key, value}}}
				break
			} else if len(current.Overflow.Entries) < isam.BlockSize {
				// if overflow block has space, then search and insert
				current.Overflow.Entries = append(current.Overflow.Entries, Record{key, value})
				break
			} else {
				// if overflow block is full, then add another overflow block to chain
				current = current.Overflow
			}
		}
	}
}

func (isam *ISAM) Search(key int) (interface{}, bool) {
	block := isam.FindBlock(key)
	i := sort.Search(len(block.Entries), func(i int) bool { return block.Entries[i].Key >= key })
	// find the i in the block entries. if ith record has the same key, then find
	if i < len(block.Entries) && block.Entries[i].Key == key {
		return block.Entries[i].Value, true
	}
	// if not , then iterate throught the overflow
	for current := block.Overflow; current != nil; current = current.Overflow {
		for _, entry := range current.Entries {
			if entry.Key == key {
				return entry.Value, true
			}
		}
	}
	return nil, false
}

func mainISAM() {
	records := []Record{{3, "a"}, {6, "b"}, {9, "c"}, {12, "d"}, {15, "e"}}
	isam := NewISAM(records, 2)
	isam.Insert(5, "f")
	val, ok := isam.Search(5)
	fmt.Println(val, ok) // Output: f true
}

type BTreeNode struct {
	leaf     bool
	keys     []int
	children []*BTreeNode
}

type BTree struct {
	root *BTreeNode
	t    int
}

func NewBTree(t int) *BTree {
	return &BTree{root: nil, t: t}
}

func (tree *BTree) Insert(key int) {
	if tree.root == nil {
		tree.root = &BTreeNode{leaf: true, keys: []int{key}}
		return
	}
	if len(tree.root.keys) >= 2*tree.t-1 {
		newRoot := &BTreeNode{children: []*BTreeNode{tree.root}}
		tree.splitChild(newRoot, 0)
		tree.root = newRoot
	}
	tree.insertNonFull(tree.root, key)
}

func (tree *BTree) insertNonFull(node *BTreeNode, key int) {
	i := len(node.keys) - 1
	if node.leaf {
		for i >= 0 && key < node.keys[i] {
			i--
		}
		node.keys = append(node.keys[:i+1], append([]int{key}, node.keys[i+1:]...)...)
	} else {
		for i >= 0 && key < node.keys[i] {
			i--
		}
		i++
		if len(node.children[i].keys) >= 2*tree.t-1 {
			tree.splitChild(node, i)
			if key > node.keys[i] {
				i++
			}
		}
		tree.insertNonFull(node.children[i], key)
	}
}

func (tree *BTree) splitChild(parent *BTreeNode, index int) {
	child := parent.children[index]
	newChild := &BTreeNode{leaf: child.leaf}
	t := tree.t

	newChild.keys = append(newChild.keys, child.keys[t:]...)
	child.keys = child.keys[:t-1]

	if !child.leaf {
		newChild.children = append(newChild.children, child.children[t:]...)
		child.children = child.children[:t]
	}

	parent.keys = append(parent.keys[:index], append([]int{child.keys[t-1]}, parent.keys[index:]...)...)
	parent.children = append(parent.children[:index+1], append([]*BTreeNode{newChild}, parent.children[index+1:]...)...)
}

func (tree *BTree) Search(key int) bool {
	return tree.search(tree.root, key)
}

func (tree *BTree) search(node *BTreeNode, key int) bool {
	if node == nil {
		return false
	}
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}
	if i < len(node.keys) && key == node.keys[i] {
		return true
	}
	if node.leaf {
		return false
	}
	return tree.search(node.children[i], key)
}

func mainB() {
	bt := NewBTree(2)
	bt.Insert(10)
	bt.Insert(20)
	bt.Insert(5)
	fmt.Println(bt.Search(5))  // Output: true
	fmt.Println(bt.Search(15)) // Output: false
}

type BPlusTreeNode struct {
	leaf     bool
	keys     []int
	children []*BPlusTreeNode
	values   []interface{}
	next     *BPlusTreeNode
}

type BPlusTree struct {
	root *BPlusTreeNode
	t    int
}

func NewBPlusTree(t int) *BPlusTree {
	return &BPlusTree{root: nil, t: t}
}

func (tree *BPlusTree) Insert(key int, value interface{}) {
	if tree.root == nil {
		tree.root = &BPlusTreeNode{leaf: true, keys: []int{key}, values: []interface{}{value}}
		return
	}
	leaf := tree.findLeafNode(key)
	tree.insertIntoLeaf(leaf, key, value)
	if len(leaf.keys) > 2*tree.t-1 {
		tree.splitLeafNode(leaf)
	}
}

func (tree *BPlusTree) findLeafNode(key int) *BPlusTreeNode {
	current := tree.root
	for !current.leaf {
		i := 0
		for i < len(current.keys) && key >= current.keys[i] {
			i++
		}
		current = current.children[i]
	}
	return current
}

func (tree *BPlusTree) insertIntoLeaf(leaf *BPlusTreeNode, key int, value interface{}) {
	i := 0
	for i < len(leaf.keys) && leaf.keys[i] < key {
		i++
	}
	leaf.keys = append(leaf.keys[:i], append([]int{key}, leaf.keys[i:]...)...)
	leaf.values = append(leaf.values[:i], append([]interface{}{value}, leaf.values[i:]...)...)
}

func (tree *BPlusTree) splitLeafNode(leaf *BPlusTreeNode) {
	t := tree.t
	newLeaf := &BPlusTreeNode{leaf: true, keys: leaf.keys[t:], values: leaf.values[t:], next: leaf.next}
	leaf.keys = leaf.keys[:t]
	leaf.values = leaf.values[:t]
	leaf.next = newLeaf
	tree.insertIntoParent(leaf, newLeaf.keys[0], newLeaf)
}

func (tree *BPlusTree) insertIntoParent(left *BPlusTreeNode, key int, right *BPlusTreeNode) {
	if left == tree.root {
		tree.root = &BPlusTreeNode{keys: []int{key}, children: []*BPlusTreeNode{left, right}}
		return
	}
	parent := tree.findParent(tree.root, left)
	i := 0
	for i < len(parent.keys) && key > parent.keys[i] {
		i++
	}
	parent.keys = append(parent.keys[:i], append([]int{key}, parent.keys[i:]...)...)
	parent.children = append(parent.children[:i+1], append([]*BPlusTreeNode{right}, parent.children[i+1:]...)...)
	if len(parent.keys) > 2*tree.t-1 {
		tree.splitInternalNode(parent)
	}
}

func (tree *BPlusTree) splitInternalNode(node *BPlusTreeNode) {
	t := tree.t
	promotedKey := node.keys[t-1]
	newNode := &BPlusTreeNode{keys: node.keys[t:], children: node.children[t:]}
	node.keys = node.keys[:t-1]
	node.children = node.children[:t]
	tree.insertIntoParent(node, promotedKey, newNode)
}

func (tree *BPlusTree) findParent(current, child *BPlusTreeNode) *BPlusTreeNode {
	if current.leaf {
		return nil
	}
	for _, c := range current.children {
		if c == child {
			return current
		}
		if parent := tree.findParent(c, child); parent != nil {
			return parent
		}
	}
	return nil
}

func (tree *BPlusTree) Search(key int) (interface{}, bool) {
	leaf := tree.findLeafNode(key)
	for i, k := range leaf.keys {
		if k == key {
			return leaf.values[i], true
		}
	}
	return nil, false
}

func (tree *BPlusTree) Delete(key int) bool {
	leaf := tree.findLeafNode(key)
	wasLeftmost := key == leaf.keys[0] // Check if deleted key was leftmost
	if !tree.deleteFromLeaf(leaf, key) {
		return false
	}
	if len(leaf.keys) < tree.t-1 {
		tree.rebalance(leaf)
	} else if wasLeftmost && len(leaf.keys) > 0 {
		// Update parent's key if leftmost key was deleted
		tree.updateParentKey(leaf, leaf.keys[0])
	}
	return true
}

func (tree *BPlusTree) deleteFromLeaf(leaf *BPlusTreeNode, key int) bool {
	i := 0
	for i < len(leaf.keys) && leaf.keys[i] < key {
		i++
	}
	if i >= len(leaf.keys) || leaf.keys[i] != key {
		return false
	}
	leaf.keys = append(leaf.keys[:i], leaf.keys[i+1:]...)
	leaf.values = append(leaf.values[:i], leaf.values[i+1:]...)
	return true
}

func (tree *BPlusTree) rebalance(node *BPlusTreeNode) {
	if node == tree.root {
		if len(node.keys) == 0 && len(node.children) > 0 {
			tree.root = node.children[0]
		}
		return
	}

	parent := tree.findParent(tree.root, node)
	index := tree.getChildIndex(parent, node)

	if index > 0 && len(parent.children[index-1].keys) > tree.t-1 {
		// Borrow from left sibling
		tree.borrowFromLeft(parent, index)
	} else if index < len(parent.children)-1 && len(parent.children[index+1].keys) > tree.t-1 {
		// Borrow from right sibling
		tree.borrowFromRight(parent, index)
	} else {
		// Merge with sibling
		if index > 0 {
			tree.mergeNodes(parent, index-1, index)
		} else {
			tree.mergeNodes(parent, index, index+1)
		}
	}
}

func (tree *BPlusTree) getChildIndex(parent, child *BPlusTreeNode) int {
	for i, c := range parent.children {
		if c == child {
			return i
		}
	}
	return -1
}

func (tree *BPlusTree) borrowFromLeft(parent *BPlusTreeNode, index int) {
	node := parent.children[index]
	leftSibling := parent.children[index-1]

	if node.leaf {
		// Move key and value
		node.keys = append([]int{leftSibling.keys[len(leftSibling.keys)-1]}, node.keys...)
		node.values = append([]interface{}{leftSibling.values[len(leftSibling.values)-1]}, node.values...)
		leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
		leftSibling.values = leftSibling.values[:len(leftSibling.values)-1]
		parent.keys[index-1] = node.keys[0]
	} else {
		// Move key and child
		node.keys = append([]int{parent.keys[index-1]}, node.keys...)
		node.children = append([]*BPlusTreeNode{leftSibling.children[len(leftSibling.children)-1]}, node.children...)
		parent.keys[index-1] = leftSibling.keys[len(leftSibling.keys)-1]
		leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
		leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
	}
}

func (tree *BPlusTree) borrowFromRight(parent *BPlusTreeNode, index int) {
	node := parent.children[index]
	rightSibling := parent.children[index+1]

	if node.leaf {
		// Move key and value
		node.keys = append(node.keys, rightSibling.keys[0])
		node.values = append(node.values, rightSibling.values[0])
		rightSibling.keys = rightSibling.keys[1:]
		rightSibling.values = rightSibling.values[1:]
		parent.keys[index] = rightSibling.keys[0]
	} else {
		// Move key and child
		node.keys = append(node.keys, parent.keys[index])
		node.children = append(node.children, rightSibling.children[0])
		parent.keys[index] = rightSibling.keys[0]
		rightSibling.keys = rightSibling.keys[1:]
		rightSibling.children = rightSibling.children[1:]
	}
}

func (tree *BPlusTree) mergeNodes(parent *BPlusTreeNode, leftIndex, rightIndex int) {
	left := parent.children[leftIndex]
	right := parent.children[rightIndex]

	if left.leaf {
		// Merge leaf nodes
		left.keys = append(left.keys, right.keys...)
		left.values = append(left.values, right.values...)
		left.next = right.next
	} else {
		// Merge internal nodes
		left.keys = append(left.keys, parent.keys[leftIndex])
		left.keys = append(left.keys, right.keys...)
		left.children = append(left.children, right.children...)
	}

	// Remove the merged node from parent
	parent.keys = append(parent.keys[:leftIndex], parent.keys[leftIndex+1:]...)
	parent.children = append(parent.children[:rightIndex], parent.children[rightIndex+1:]...)

	if len(parent.keys) < tree.t-1 {
		tree.rebalance(parent)
	}
}

func (tree *BPlusTree) updateParentKey(node *BPlusTreeNode, newKey int) {
	parent := tree.findParent(tree.root, node)
	if parent == nil {
		return
	}

	// Find the index of the child pointer to this node
	childIndex := -1
	for i, child := range parent.children {
		if child == node {
			childIndex = i
			break
		}
	}

	// If this is not the first child, update the corresponding key
	if childIndex > 0 {
		parent.keys[childIndex-1] = newKey
	}
}

func mainBplus() {
	bpt := NewBPlusTree(2)
	bpt.Insert(10, "a")
	bpt.Insert(20, "b")
	bpt.Insert(5, "c")
	val, ok := bpt.Search(5)
	fmt.Println(val, ok) // Output: c true
}
