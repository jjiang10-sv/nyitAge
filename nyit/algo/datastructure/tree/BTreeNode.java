package tree;

import java.util.ArrayList;
import java.util.List;

public // B-Tree Node class
class BTreeNode {
    boolean isLeaf; // Whether this node is a leaf
    List<Integer> keys; // List of keys in the node
    List<BTreeNode> children; // List of child nodes

    public BTreeNode(boolean isLeaf) {
        this.isLeaf = isLeaf;
        this.keys = new ArrayList<>();
        this.children = new ArrayList<>();
    }
}

// B-Tree class
class BTree {
    private BTreeNode root; // Root node of the tree
    private int t; // Minimum degree (defines the range for number of keys)

    // Constructor
    public BTree(int t) {
        this.root = null;
        this.t = t;
    }

    // Insert a key into the tree
    public void insert(int key) {
        if (root == null) {
            root = new BTreeNode(true);
            root.keys.add(key);
            return;
        }

        // If root is full, split it
        if (root.keys.size() == 2 * t - 1) {
            BTreeNode newRoot = new BTreeNode(false);
            newRoot.children.add(root);
            splitChild(newRoot, 0);
            root = newRoot;
        }
        insertNonFull(root, key);
    }

    // Insert a key into a non-full node
    private void insertNonFull(BTreeNode node, int key) {
        int i = node.keys.size() - 1;

        if (node.isLeaf) {
            // Find the correct position to insert the key
            while (i >= 0 && key < node.keys.get(i)) {
                i--;
            }
            node.keys.add(i + 1, key);
        } else {
            // Find the appropriate child to follow
            while (i >= 0 && key < node.keys.get(i)) {
                i--;
            }
            i++;

            // If the child is full, split it
            if (node.children.get(i).keys.size() == 2 * t - 1) {
                splitChild(node, i);
                if (key > node.keys.get(i)) {
                    i++;
                }
            }
            insertNonFull(node.children.get(i), key);
        }
    }

    // Split a full child node
    private void splitChild(BTreeNode parent, int index) {
        BTreeNode child = parent.children.get(index);
        BTreeNode newChild = new BTreeNode(child.isLeaf);

        // Move the second half of keys to the new child
        newChild.keys.addAll(child.keys.subList(t, child.keys.size()));
        child.keys.subList(t - 1, child.keys.size()).clear();

        // If not a leaf, move the second half of children
        if (!child.isLeaf) {
            newChild.children.addAll(child.children.subList(t, child.children.size()));
            child.children.subList(t, child.children.size()).clear();
        }

        // Insert the middle key into the parent
        parent.keys.add(index, child.keys.get(t - 1));
        parent.children.add(index + 1, newChild);
    }

    // Search for a key in the tree
    public boolean search(int key) {
        return search(root, key);
    }

    // Recursive search helper
    private boolean search(BTreeNode node, int key) {
        if (node == null) {
            return false;
        }

        int i = 0;
        while (i < node.keys.size() && key > node.keys.get(i)) {
            i++;
        }

        if (i < node.keys.size() && key == node.keys.get(i)) {
            return true;
        }

        if (node.isLeaf) {
            return false;
        }

        return search(node.children.get(i), key);
    }

    // Main method for testing
    public static void main(String[] args) {
        BTree bt = new BTree(2);
        bt.insert(10);
        bt.insert(20);
        bt.insert(5);
        System.out.println(bt.search(5));  // Output: true
        System.out.println(bt.search(15)); // Output: false
    }
}
