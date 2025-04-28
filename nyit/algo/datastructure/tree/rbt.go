// Here's a comprehensive implementation of a Red-Black Tree in Go, covering insertion and deletion with all necessary cases:

// ```go
package tree

import "fmt"

type Color bool

const (
    Red   Color = true
    Black Color = false
)

type Node struct {
    key    int
    color  Color
    left   *Node
    right  *Node
    parent *Node
}

type RedBlackTree struct {
    root    *Node
    nilNode *Node // Sentinel node
}

func NewRedBlackTree() *RedBlackTree {
    nilNode := &Node{color: Black}
    return &RedBlackTree{
        root:    nilNode,
        nilNode: nilNode,
    }
}

func (t *RedBlackTree) Insert(key int) {
    newNode := &Node{
        key:    key,
        color:  Red,
        left:   t.nilNode,
        right:  t.nilNode,
        parent: t.nilNode,
    }

    parent := t.nilNode
    current := t.root
    for current != t.nilNode {
        parent = current
        if newNode.key < current.key {
            current = current.left
        } else {
            current = current.right
        }
    }

    newNode.parent = parent
    if parent == t.nilNode {
        t.root = newNode
    } else if newNode.key < parent.key {
        parent.left = newNode
    } else {
        parent.right = newNode
    }

    t.fixInsert(newNode)
}

func (t *RedBlackTree) fixInsert(node *Node) {
    for node.parent.color == Red {
        if node.parent == node.parent.parent.left {
            uncle := node.parent.parent.right
            if uncle.color == Red {
                // Case 1: Recolor
                node.parent.color = Black
                uncle.color = Black
                node.parent.parent.color = Red
                node = node.parent.parent
            } else {
                if node == node.parent.right {
                    // Case 2: Left-Rotate
                    node = node.parent
                    t.leftRotate(node)
                }
                // Case 3: Recolor and Right-Rotate
                node.parent.color = Black
                node.parent.parent.color = Red
                t.rightRotate(node.parent.parent)
            }
        } else {
            uncle := node.parent.parent.left
            if uncle.color == Red {
                // Case 1: Recolor
                node.parent.color = Black
                uncle.color = Black
                node.parent.parent.color = Red
                node = node.parent.parent
            } else {
                if node == node.parent.left {
                    // Case 2: Right-Rotate
                    node = node.parent
                    t.rightRotate(node)
                }
                // Case 3: Recolor and Left-Rotate
                node.parent.color = Black
                node.parent.parent.color = Red
                t.leftRotate(node.parent.parent)
            }
        }
    }
    t.root.color = Black
}

func (t *RedBlackTree) leftRotate(x *Node) {
    y := x.right
    x.right = y.left
    if y.left != t.nilNode {
        y.left.parent = x
    }
    y.parent = x.parent
    if x.parent == t.nilNode {
        t.root = y
    } else if x == x.parent.left {
        x.parent.left = y
    } else {
        x.parent.right = y
    }
    y.left = x
    x.parent = y
}

func (t *RedBlackTree) rightRotate(y *Node) {
    x := y.left
    y.left = x.right
    if x.right != t.nilNode {
        x.right.parent = y
    }
    x.parent = y.parent
    if y.parent == t.nilNode {
        t.root = x
    } else if y == y.parent.right {
        y.parent.right = x
    } else {
        y.parent.left = x
    }
    x.right = y
    y.parent = x
}

func (t *RedBlackTree) Delete(key int) {
    z := t.findNode(key)
    if z == t.nilNode {
        return
    }

    y := z
    yOriginalColor := y.color
    var x *Node

    if z.left == t.nilNode {
        x = z.right
        t.transplant(z, z.right)
    } else if z.right == t.nilNode {
        x = z.left
        t.transplant(z, z.left)
    } else {
        y = t.minimum(z.right)
        yOriginalColor = y.color
        x = y.right
        if y.parent == z {
            x.parent = y
        } else {
            t.transplant(y, y.right)
            y.right = z.right
            y.right.parent = y
        }
        t.transplant(z, y)
        y.left = z.left
        y.left.parent = y
        y.color = z.color
    }

    if yOriginalColor == Black {
        t.fixDelete(x)
    }
}

func (t *RedBlackTree) transplant(u, v *Node) {
    if u.parent == t.nilNode {
        t.root = v
    } else if u == u.parent.left {
        u.parent.left = v
    } else {
        u.parent.right = v
    }
    v.parent = u.parent
}

func (t *RedBlackTree) fixDelete(x *Node) {
    for x != t.root && x.color == Black {
        if x == x.parent.left {
            s := x.parent.right
            if s.color == Red {
                // Case 1: Sibling is red
                s.color = Black
                x.parent.color = Red
                t.leftRotate(x.parent)
                s = x.parent.right
            }

            if s.left.color == Black && s.right.color == Black {
                // Case 2: Both children black
                s.color = Red
                x = x.parent
            } else {
                if s.right.color == Black {
                    // Case 3: Right child black
                    s.left.color = Black
                    s.color = Red
                    t.rightRotate(s)
                    s = x.parent.right
                }
                // Case 4: Right child red
                s.color = x.parent.color
                x.parent.color = Black
                s.right.color = Black
                t.leftRotate(x.parent)
                x = t.root
            }
        } else {
            s := x.parent.left
            if s.color == Red {
                // Case 1: Sibling is red
                s.color = Black
                x.parent.color = Red
                t.rightRotate(x.parent)
                s = x.parent.left
            }

            if s.right.color == Black && s.left.color == Black {
                // Case 2: Both children black
                s.color = Red
                x = x.parent
            } else {
                if s.left.color == Black {
                    // Case 3: Left child black
                    s.right.color = Black
                    s.color = Red
                    t.leftRotate(s)
                    s = x.parent.left
                }
                // Case 4: Left child red
                s.color = x.parent.color
                x.parent.color = Black
                s.left.color = Black
                t.rightRotate(x.parent)
                x = t.root
            }
        }
    }
    x.color = Black
}

func (t *RedBlackTree) findNode(key int) *Node {
    current := t.root
    for current != t.nilNode {
        if key == current.key {
            return current
        } else if key < current.key {
            current = current.left
        } else {
            current = current.right
        }
    }
    return t.nilNode
}

func (t *RedBlackTree) minimum(node *Node) *Node {
    for node.left != t.nilNode {
        node = node.left
    }
    return node
}

// Helper function to print the tree (in-order traversal)
func (t *RedBlackTree) PrintInOrder(node *Node) {
    if node != t.nilNode {
        t.PrintInOrder(node.left)
        color := "B"
        if node.color == Red {
            color = "R"
        }
        fmt.Printf("%d%s ", node.key, color)
        t.PrintInOrder(node.right)
    }
}

func mainRbt() {
    rbt := NewRedBlackTree()
    keys := []int{7, 3, 18, 10, 22, 8, 11, 26}
    for _, key := range keys {
        rbt.Insert(key)
    }

    fmt.Println("In-order traversal after insertions:")
    rbt.PrintInOrder(rbt.root)
    fmt.Println()

    rbt.Delete(18)
    rbt.Delete(11)
    rbt.Delete(3)

    fmt.Println("In-order traversal after deletions:")
    rbt.PrintInOrder(rbt.root)
    fmt.Println()
}
// ```

// This implementation includes:

// 1. **Node Structure**: Each node has a key, color, and pointers to left/right children and parent.
// 2. **Insertion**: Follows standard BST insertion and fixes violations using rotations and recoloring.
// 3. **Deletion**: Handles all cases including sibling color checks and rotations.
// 4. **Rotations**: Left and right rotations to maintain tree balance.
// 5. **Fixup Functions**: `fixInsert` and `fixDelete` handle all necessary cases for maintaining Red-Black properties.
// 6. **Helper Functions**: For tree operations like transplant, minimum node, and node search.

// Key features:
// - Uses a sentinel node for nil leaves
// - Proper handling of all insertion and deletion cases
// - Color flipping and rotations to maintain balance
// - In-order traversal print function for verification

// The main function demonstrates insertion of several keys followed by deletions, with in-order traversal output showing the tree structure and node colors.