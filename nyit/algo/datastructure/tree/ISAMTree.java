package tree;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

// Record class to store key-value pairs
class Record {
    int key;
    Object value;

    public Record(int key, Object value) {
        this.key = key;
        this.value = value;
    }
}

// Block class to store records and overflow blocks
class Block {
    List<Record> entries; // List of records in the block
    Block overflow; // Pointer to overflow block

    public Block() {
        this.entries = new ArrayList<>();
        this.overflow = null;
    }
}

// ISAM Tree class
class ISAMTree {
    private List<Integer> index; // Sorted max keys of each primary block
    private List<Block> primary; // List of primary blocks
    private int blockSize; // Maximum number of records per block

    // Constructor
    public ISAMTree(List<Record> records, int blockSize) {
        this.blockSize = blockSize;
        this.index = new ArrayList<>();
        this.primary = new ArrayList<>();

        // Sort records by key
        records.sort((r1, r2) -> Integer.compare(r1.key, r2.key));

        // Create primary blocks
        for (int i = 0; i < records.size(); i += blockSize) {
            int end = Math.min(i + blockSize, records.size());
            Block block = new Block();
            block.entries.addAll(records.subList(i, end));
            primary.add(block);

            if (!block.entries.isEmpty()) {
                index.add(block.entries.get(block.entries.size() - 1).key);
            }
        }
    }

    // Find the block where a key should be located
    private Block findBlock(int key) {
        int i = Collections.binarySearch(index, key);
        if (i < 0) {
            i = -(i + 1);
        }
        if (i >= index.size()) {
            i = index.size() - 1;
        }
        return primary.get(i);
    }

    // Insert a key-value pair into the tree
    public void insert(int key, Object value) {
        Block block = findBlock(key);

        if (block.entries.size() < blockSize) {
            // Insert into primary block if there's space
            int i = Collections.binarySearch(
                block.entries.stream().map(r -> r.key).toList(),
                key
            );
            if (i < 0) {
                i = -(i + 1);
            }
            block.entries.add(i, new Record(key, value));
        } else {
            // Insert into overflow block
            Block current = block;
            while (true) {
                if (current.overflow == null) {
                    current.overflow = new Block();
                    current.overflow.entries.add(new Record(key, value));
                    break;
                } else if (current.overflow.entries.size() < blockSize) {
                    current.overflow.entries.add(new Record(key, value));
                    break;
                } else {
                    current = current.overflow;
                }
            }
        }
    }

    // Search for a key in the tree
    public Object search(int key) {
        Block block = findBlock(key);

        // Search in primary block
        for (Record record : block.entries) {
            if (record.key == key) {
                return record.value;
            }
        }

        // Search in overflow blocks
        Block current = block.overflow;
        while (current != null) {
            for (Record record : current.entries) {
                if (record.key == key) {
                    return record.value;
                }
            }
            current = current.overflow;
        }

        return null;
    }

    // Main method for testing
    public static void main(String[] args) {
        List<Record> records = new ArrayList<>();
        records.add(new Record(3, "a"));
        records.add(new Record(6, "b"));
        records.add(new Record(9, "c"));
        records.add(new Record(12, "d"));
        records.add(new Record(15, "e"));

        ISAMTree isam = new ISAMTree(records, 2);
        isam.insert(5, "f");
        Object val = isam.search(5);
        System.out.println(val); // Output: f
    }
} 

// Here are some reliable reference links for B-Tree, B+ Tree, and ISAM Tree concepts:

// ### B-Tree References:
// 1. **GeeksforGeeks - B-Tree Introduction**  
//    [https://www.geeksforgeeks.org/introduction-of-b-tree-2/](https://www.geeksforgeeks.org/introduction-of-b-tree-2/)  
//    - Covers the basics of B-Tree, its properties, and operations.

// 2. **Wikipedia - B-Tree**  
//    [https://en.wikipedia.org/wiki/B-tree](https://en.wikipedia.org/wiki/B-tree)  
//    - Detailed explanation of B-Tree structure, algorithms, and applications.

// 3. **TutorialsPoint - B-Tree**  
//    [https://www.tutorialspoint.com/data_structures_algorithms/b_tree.htm](https://www.tutorialspoint.com/data_structures_algorithms/b_tree.htm)  
//    - Step-by-step guide to B-Tree operations with examples.

// ---

// ### B+ Tree References:
// 1. **GeeksforGeeks - B+ Tree Introduction**  
//    [https://www.geeksforgeeks.org/introduction-of-b-tree/](https://www.geeksforgeeks.org/introduction-of-b-tree/)  
//    - Explains the structure and advantages of B+ Trees over B-Trees.

// 2. **Wikipedia - B+ Tree**  
//    [https://en.wikipedia.org/wiki/B%2B_tree](https://en.wikipedia.org/wiki/B%2B_tree)  
//    - Comprehensive overview of B+ Tree properties and use cases.

// 3. **Javatpoint - B+ Tree**  
//    [https://www.javatpoint.com/b-plus-tree](https://www.javatpoint.com/b-plus-tree)  
//    - Simple and clear explanation of B+ Tree operations.

// ---

// ### ISAM Tree References:
// 1. **Wikipedia - Indexed Sequential Access Method (ISAM)**  
//    [https://en.wikipedia.org/wiki/ISAM](https://en.wikipedia.org/wiki/ISAM)  
//    - Overview of ISAM, its structure, and applications in databases.

// 2. **GeeksforGeeks - ISAM (Indexed Sequential Access Method)**  
//    [https://www.geeksforgeeks.org/isam-indexed-sequential-access-method/](https://www.geeksforgeeks.org/isam-indexed-sequential-access-method/)  
//    - Explanation of ISAM with a focus on its use in file organization.

// 3. **TutorialsPoint - ISAM**  
//    [https://www.tutorialspoint.com/dbms/isam.htm](https://www.tutorialspoint.com/dbms/isam.htm)  
//    - Covers the basics of ISAM and its advantages in database systems.

// ---

// ### Additional Resources:
// 1. **CMU Database Systems - Indexing**  
//    [https://15445.courses.cs.cmu.edu/fall2022/notes/06-indexing.pdf](https://15445.courses.cs.cmu.edu/fall2022/notes/06-indexing.pdf)  
//    - Lecture notes from Carnegie Mellon University on indexing, including B-Trees and B+ Trees.

// 2. **Stanford Database Systems - Indexing**  
//    [https://cs145-fa22.github.io/notes/06-indexing.pdf](https://cs145-fa22.github.io/notes/06-indexing.pdf)  
//    - Detailed notes on indexing techniques from Stanford University.

// These links provide a mix of theoretical explanations, practical examples, and visualizations to help you understand these tree structures better. Let me know if you need more specific resources!
