package tree

import (
	"fmt"
	"sort"
	"strings"
)

// BTreeNode represents a node in the B-Tree
type BTreeNode struct {
	leaf     bool
	keys     []string
	values   [][]byte
	children []*BTreeNode
}

// BTree represents the whole B-Tree
type BTree struct {
	root      *BTreeNode
	minDegree int // t
}

// NewBTree creates an empty B-tree with min degree t (t >= 2)
func NewBTree(t int) *BTree {
	if t < 2 {
		panic("minDegree t must be >= 2")
	}
	n := &BTreeNode{leaf: true, keys: []string{}, values: [][]byte{}, children: []*BTreeNode{}}
	return &BTree{root: n, minDegree: t}
}

/* ------------------------
   Search
   ------------------------ */

// Search returns (value, found)
func (t *BTree) Search(key string) ([]byte, bool) {
	return t.root.search(key)
}

func (n *BTreeNode) search(key string) ([]byte, bool) {
	// find first index i such that keys[i] >= key
	i := sort.Search(len(n.keys), func(i int) bool { return n.keys[i] >= key })
	if i < len(n.keys) && n.keys[i] == key {
		return n.values[i], true
	}
	if n.leaf {
		return nil, false
	}
	return n.children[i].search(key)
}

/* ------------------------
   Traversal / List
   ------------------------ */

// Traverse returns all key-value pairs in sorted order
func (t *BTree) Traverse() []KV {
	out := []KV{}
	t.root.traverse(&out)
	return out
}

// type KV struct {
// 	Key   string
// 	Value []byte
// }

func (n *BTreeNode) traverse(out *[]KV) {
	for i := 0; i < len(n.keys); i++ {
		if !n.leaf {
			n.children[i].traverse(out)
		}
		// each key shows once in leaf or in the internal node.
		*out = append(*out, KV{Key: n.keys[i], Value: n.values[i]})

	}
	if !n.leaf {
		if len(n.children) == len(n.keys)+1 {
			n.children[len(n.keys)].traverse(out)
		}

	}
}

// RangeList returns key-values with start <= key < end (empty string means unbounded)
func (t *BTree) RangeList(start, end string) []KV {
	all := t.Traverse()
	out := []KV{}
	for _, kv := range all {
		if (start == "" || kv.Key >= start) && (end == "" || kv.Key < end) {
			out = append(out, kv)
		}
	}
	return out
}

/* ------------------------
   Insert
   ------------------------ */

// Insert or Update
func (t *BTree) Insert(key string, value []byte) {
	r := t.root
	if len(r.keys) == 2*t.minDegree-1 {
		// root is full -> split
		s := &BTreeNode{leaf: false, keys: []string{}, values: [][]byte{}, children: []*BTreeNode{r}}
		t.root = s
		t.splitChild(s, 0)
		t.insertNonFull(s, key, value)
	} else {
		t.insertNonFull(r, key, value)
	}
}

func (t *BTree) insertNonFull(n *BTreeNode, key string, value []byte) {
	i := len(n.keys) - 1
	if n.leaf {
		// insert in leaf (keep sorted order) - if key exists, replace value
		pos := sort.Search(len(n.keys), func(i int) bool { return n.keys[i] >= key })
		if pos < len(n.keys) && n.keys[pos] == key {
			n.values[pos] = value
			return
		}
		// insert
		n.keys = append(n.keys, "")      // extend
		n.values = append(n.values, nil) // extend
		copy(n.keys[pos+1:], n.keys[pos:])
		copy(n.values[pos+1:], n.values[pos:])
		n.keys[pos] = key
		n.values[pos] = value
		return
	}
	// internal node: find child to descend into
	for i >= 0 && key < n.keys[i] {
		i--
	}
	i++
	// if child full, split then decide which of the two to descend
	if len(n.children[i].keys) == 2*t.minDegree-1 {
		t.splitChild(n, i)
		if key > n.keys[i] {
			i++
		} else if key == n.keys[i] {
			// if equal to promoted key, update value in parent
			n.values[i] = value
			return
		}
	}
	t.insertNonFull(n.children[i], key, value)
}

// a potential bug : what if the parent keys length reached 2*t.minDegree-1, then the parent need split
func (t *BTree) splitChild(parent *BTreeNode, idx int) {
	tNode := parent.children[idx]
	tDeg := t.minDegree
	// need get the k-v which will move up after the split
	moveUpKey := tNode.keys[tDeg-1]
	moveUpval := tNode.values[tDeg-1]

	// new node z will hold t-1 keys from y
	z := &BTreeNode{leaf: tNode.leaf}
	// copy last t-1 keys/values to z
	z.keys = append(z.keys, tNode.keys[tDeg:]...)
	z.values = append(z.values, tNode.values[tDeg:]...)
	// if not leaf, copy last t children
	if !tNode.leaf {
		z.children = append(z.children, tNode.children[tDeg:]...)
	}
	// shrink tNode
	tNode.keys = tNode.keys[:tDeg-1]
	tNode.values = tNode.values[:tDeg-1]
	if !tNode.leaf {
		tNode.children = tNode.children[:tDeg]
	}
	// insert z as child of parent
	parent.children = append(parent.children, nil)
	copy(parent.children[idx+2:], parent.children[idx+1:])
	parent.children[idx+1] = z

	// move middle key up to parent
	parent.keys = append(parent.keys, "")
	parent.values = append(parent.values, nil)
	copy(parent.keys[idx+1:], parent.keys[idx:])
	copy(parent.values[idx+1:], parent.values[idx:])
	parent.keys[idx] = moveUpKey
	parent.values[idx] = moveUpval
	if len(parent.keys) == 2*t.minDegree-1 {
		fmt.Println("need split the parent")
	}
	// Note: tNode.keys already shrunk above so the mid key was preserved prior to shrink;
	// to avoid confusion we stored mid key before truncation in practice. For clarity we adjust:
	// The above works because we copied slices before shrinking; in Go the above order is safe.
}

/* ------------------------
   Delete (CLRS algorithm)
   ------------------------ */

// Delete removes a key if exists
func (t *BTree) Delete(key string) {
	t.delete(t.root, key)
	// shrink root if it has 0 keys and is non-leaf
	if len(t.root.keys) == 0 && !t.root.leaf {
		t.root = t.root.children[0]
	}
}

func (t *BTree) delete(n *BTreeNode, key string) {
	idx := sort.Search(len(n.keys), func(i int) bool { return n.keys[i] >= key })
	// case 1: key present in this node
	if idx < len(n.keys) && n.keys[idx] == key {
		if n.leaf {
			// case 1a: leaf node -> remove key directly
			n.keys = append(n.keys[:idx], n.keys[idx+1:]...)
			n.values = append(n.values[:idx], n.values[idx+1:]...)
			return
		}
		// case 1b: internal node
		left := n.children[idx]
		right := n.children[idx+1]
		if len(left.keys) >= t.minDegree {
			// predecessor
			predKey, predVal := t.getPredecessor(left)
			n.keys[idx] = predKey
			n.values[idx] = predVal
			t.delete(left, predKey)
			return
		} else if len(right.keys) >= t.minDegree {
			// successor
			succKey, succVal := t.getSuccessor(right)
			n.keys[idx] = succKey
			n.values[idx] = succVal
			t.delete(right, succKey)
			return
		} else {
			// merge key and right into left
			t.merge(n, idx)
			t.delete(left, key)
			return
		}
	}
	// case 2: key not present in this node
	if n.leaf {
		// not found
		return
	}
	// ensure child idx has at least t keys before descending
	child := n.children[idx]
	if len(child.keys) < t.minDegree {
		// try to borrow from siblings or merge
		if idx > 0 && len(n.children[idx-1].keys) >= t.minDegree {
			t.borrowFromPrev(n, idx)
		} else if idx < len(n.children)-1 && len(n.children[idx+1].keys) >= t.minDegree {
			t.borrowFromNext(n, idx)
		} else {
			// merge with sibling
			if idx < len(n.children)-1 {
				t.merge(n, idx)
			} else {
				t.merge(n, idx-1)
				child = n.children[idx-1]
			}
		}
	}
	// descend
	t.delete(child, key)
}

func (t *BTree) getPredecessor(n *BTreeNode) (string, []byte) {
	cur := n
	for !cur.leaf {
		cur = cur.children[len(cur.children)-1]
	}
	return cur.keys[len(cur.keys)-1], cur.values[len(cur.values)-1]
}

func (t *BTree) getSuccessor(n *BTreeNode) (string, []byte) {
	cur := n
	for !cur.leaf {
		cur = cur.children[0]
	}
	return cur.keys[0], cur.values[0]
}

// merge child idx and idx+1 into child idx
func (t *BTree) merge(parent *BTreeNode, idx int) {
	child := parent.children[idx]
	sibling := parent.children[idx+1]
	// pull down parent key
	child.keys = append(child.keys, parent.keys[idx])
	child.values = append(child.values, parent.values[idx])
	// append sibling keys/values
	child.keys = append(child.keys, sibling.keys...)
	child.values = append(child.values, sibling.values...)
	// append sibling children if any
	if !child.leaf {
		child.children = append(child.children, sibling.children...)
	}
	// remove parent key and sibling pointer
	parent.keys = append(parent.keys[:idx], parent.keys[idx+1:]...)
	parent.values = append(parent.values[:idx], parent.values[idx+1:]...)
	parent.children = append(parent.children[:idx+1], parent.children[idx+2:]...)
}

// borrow from children[idx-1] -> children[idx]
func (t *BTree) borrowFromPrev(parent *BTreeNode, idx int) {
	child := parent.children[idx]
	sibling := parent.children[idx-1]

	// shift child's keys right
	child.keys = append([]string{parent.keys[idx-1]}, child.keys...)
	child.values = append([][]byte{parent.values[idx-1]}, child.values...)
	if !child.leaf {
		child.children = append([]*BTreeNode{sibling.children[len(sibling.children)-1]}, child.children...)
		sibling.children = sibling.children[:len(sibling.children)-1]
	}

	// move last key of sibling up to parent
	parent.keys[idx-1] = sibling.keys[len(sibling.keys)-1]
	parent.values[idx-1] = sibling.values[len(sibling.values)-1]
	sibling.keys = sibling.keys[:len(sibling.keys)-1]
	sibling.values = sibling.values[:len(sibling.values)-1]
}

// borrow from children[idx+1] -> children[idx]
func (t *BTree) borrowFromNext(parent *BTreeNode, idx int) {
	child := parent.children[idx]
	sibling := parent.children[idx+1]

	// move parent key into child
	child.keys = append(child.keys, parent.keys[idx])
	child.values = append(child.values, parent.values[idx])
	if !child.leaf {
		child.children = append(child.children, sibling.children[0])
		sibling.children = sibling.children[1:]
	}

	// move sibling first key up to parent
	parent.keys[idx] = sibling.keys[0]
	parent.values[idx] = sibling.values[0]
	sibling.keys = sibling.keys[1:]
	sibling.values = sibling.values[1:]
}

/* ------------------------
   Helpers & demo
   ------------------------ */

func (t *BTree) Dump() {
	fmt.Println("BTree dump:")
	t.root.dump(0)
}

func (n *BTreeNode) dump(level int) {
	fmt.Printf("%sNode (leaf=%v) keys=%v\n", indent(level), n.leaf, n.keys)
	for _, c := range n.children {
		c.dump(level + 1)
	}
}

func indent(n int) string { return strings.Repeat(" ", n) }
