package tree

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

type KVBPlus struct {
	key   int
	value interface{}
}

func (tree *BPlusTree) Traversal() *[]KVBPlus {
	out := &[]KVBPlus{}
	tree.root.traversal(out)
	return out
}

func (node *BPlusTreeNode) traversal(out *[]KVBPlus) {
	i := 0
	for i < len(node.keys) {
		child := node.children[i]
		child.traversal_(out)
		i++
	}
	if len(node.children) > len(node.keys) {
		child := node.children[i]
		child.traversal_(out)
	}
}

func (node *BPlusTreeNode) traversal_(out *[]KVBPlus) {

	if !node.leaf {
		node.traversal(out)
	} else {
		for i, k := range node.keys {
			*out = append(*out, KVBPlus{
				key:   k,
				value: node.values[i],
			})
		}
	}
}
