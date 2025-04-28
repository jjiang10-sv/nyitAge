package tree;

import java.util.ArrayList;
import java.util.List;
class BPlusTreeNode {
    boolean isLeaf; // Whether this node is a leaf
    List<Integer> keys; // List of keys in the node
    List<BPlusTreeNode> children; // List of child nodes (for non-leaf nodes)
    List<Object> values; // List of values (for leaf nodes)
    BPlusTreeNode next; // Pointer to next leaf node (for leaf nodes)

    public BPlusTreeNode(boolean isLeaf) {
        this.isLeaf = isLeaf;
        this.keys = new ArrayList<>();
        if (isLeaf) {
            this.values = new ArrayList<>();
        } else {
            this.children = new ArrayList<>();
        }
    }
}

// B+ Tree class
class BPlusTree {
    private BPlusTreeNode root; // Root node of the tree
    private int t; // Minimum degree (defines the range for number of keys)

    // Constructor
    public BPlusTree(int t) {
        this.root = null;
        this.t = t;
    }

    // Insert a key-value pair into the tree
    public void insert(int key, Object value) {
        if (root == null) {
            root = new BPlusTreeNode(true);
            root.keys.add(key);
            root.values.add(value);
            return;
        }
        
        BPlusTreeNode leaf = findLeafNode(key);
        insertIntoLeaf(leaf, key, value);
        
        // Split leaf node if it exceeds maximum capacity
        if (leaf.keys.size() > 2 * t - 1) {
            splitLeafNode(leaf);
        }
    }

    // Find the leaf node where the key should be inserted
    private BPlusTreeNode findLeafNode(int key) {
        BPlusTreeNode current = root;
        while (!current.isLeaf) {
            int i = 0;
            // Find the appropriate child to follow
            while (i < current.keys.size() && key >= current.keys.get(i)) {
                i++;
            }
            current = current.children.get(i);
        }
        return current;
    }

    // Insert a key-value pair into a leaf node
    private void insertIntoLeaf(BPlusTreeNode leaf, int key, Object value) {
        int i = 0;
        // Find the correct position to insert the key
        while (i < leaf.keys.size() && leaf.keys.get(i) < key) {
            i++;
        }
        leaf.keys.add(i, key);
        leaf.values.add(i, value);
    }

    // Split an overflowing leaf node
    private void splitLeafNode(BPlusTreeNode leaf) {
        BPlusTreeNode newLeaf = new BPlusTreeNode(true);
        // Move the second half of keys and values to the new leaf
        newLeaf.keys.addAll(leaf.keys.subList(t, leaf.keys.size()));
        newLeaf.values.addAll(leaf.values.subList(t, leaf.values.size()));
        leaf.keys.subList(t, leaf.keys.size()).clear();
        leaf.values.subList(t, leaf.values.size()).clear();
        
        // Update the linked list of leaf nodes
        newLeaf.next = leaf.next;
        leaf.next = newLeaf;
        
        // Insert the new leaf into the parent node
        insertIntoParent(leaf, newLeaf.keys.get(0), newLeaf);
    }

    // Insert a new child into the parent node
    private void insertIntoParent(BPlusTreeNode left, int key, BPlusTreeNode right) {
        if (left == root) {
            // Create a new root if we're splitting the root
            root = new BPlusTreeNode(false);
            root.keys.add(key);
            root.children.add(left);
            root.children.add(right);
            return;
        }
        
        BPlusTreeNode parent = findParent(root, left);
        int i = 0;
        // Find the correct position to insert the new key
        while (i < parent.keys.size() && key > parent.keys.get(i)) {
            i++;
        }
        parent.keys.add(i, key);
        parent.children.add(i + 1, right);
        
        // Split the parent node if it's now too large
        if (parent.keys.size() > 2 * t - 1) {
            splitInternalNode(parent);
        }
    }

    // Split an overflowing internal node
    private void splitInternalNode(BPlusTreeNode node) {
        int promotedKey = node.keys.get(t - 1);
        BPlusTreeNode newNode = new BPlusTreeNode(false);
        // Move the second half of keys and children to the new node
        newNode.keys.addAll(node.keys.subList(t, node.keys.size()));
        newNode.children.addAll(node.children.subList(t, node.children.size()));
        node.keys.subList(t - 1, node.keys.size()).clear();
        node.children.subList(t, node.children.size()).clear();
        
        // Insert the new node into the parent
        insertIntoParent(node, promotedKey, newNode);
    }

    // Find the parent of a given node
    private BPlusTreeNode findParent(BPlusTreeNode current, BPlusTreeNode child) {
        if (current.isLeaf) {
            return null;
        }
        for (int i = 0; i < current.children.size(); i++) {
            if (current.children.get(i) == child) {
                return current;
            }
            BPlusTreeNode parent = findParent(current.children.get(i), child);
            if (parent != null) {
                return parent;
            }
        }
        return null;
    }

    // Search for a key in the tree
    public Object search(int key) {
        BPlusTreeNode leaf = findLeafNode(key);
        for (int i = 0; i < leaf.keys.size(); i++) {
            if (leaf.keys.get(i) == key) {
                return leaf.values.get(i);
            }
        }
        return null;
    }

    // Main method for testing
    public static void main(String[] args) {
        BPlusTree bpt = new BPlusTree(2);
        bpt.insert(10, "a");
        bpt.insert(20, "b");
        bpt.insert(5, "c");
        Object val = bpt.search(5);
        System.out.println(val); // Output: c
    }
} 