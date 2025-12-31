package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//// Example usage ////

func TestISMTree(t *testing.T) {
	dir := "./data_lsm_example"
	os.RemoveAll(dir)
	db, err := Open(dir, 1024) // flush threshold 1024 bytes
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Putting keys...")
	_ = db.Put("apple", []byte("fruit"))
	_ = db.Put("banana", []byte("yellow"))
	_ = db.Put("carrot", []byte("root"))
	_ = db.Put("apricot", []byte("fruit2"))

	fmt.Println("Get banana:")
	if v, ok, _ := db.Get("banana"); ok {
		fmt.Println("banana ->", string(v))
	}

	fmt.Println("List a..z")
	list, _ := db.List("a", "z")
	for _, kv := range list {
		fmt.Printf("%s => %s\n", kv.Key, string(kv.Value))
	}

	fmt.Println("Delete banana")
	_ = db.Delete("banana")
	if v, ok, _ := db.Get("banana"); !ok {
		fmt.Println("banana deleted")
	} else {
		fmt.Println("banana ->", string(v))
	}

	fmt.Println("Force flush and compaction")
	_ = db.Flush()
	_ = db.Compact()

	fmt.Println("List after compaction")
	list, _ = db.List("", "")
	for _, kv := range list {
		fmt.Printf("%s => %s\n", kv.Key, string(kv.Value))
	}

	// show sstable files
	files, _ := filepath.Glob(filepath.Join(dir, "sst_*.sst"))
	fmt.Println("SSTables:", strings.Join(files, ", "))
}

func TestISAM(t *testing.T) {
	records := []Record{{3, "a"}, {6, "b"}, {9, "c"}, {12, "d"}, {15, "e"}}
	isam := NewISAM(records, 2)
	isam.Insert(5, "f")
	val, ok := isam.Search(5)
	fmt.Println(val, ok) // Output: f true
}

func TestBTree(t *testing.T) {
	bt := NewBTree(2) // minDegree t=2 (max keys = 3)
	entries := []struct {
		k string
		v string
	}{
		{"G", "7"}, {"M", "13"}, {"P", "16"}, {"X", "24"},
		{"A", "1"}, {"C", "3"}, {"D", "4"}, {"E", "5"},
		{"J", "10"}, {"K", "11"}, {"N", "14"}, {"O", "15"},
		{"R", "18"}, {"S", "19"}, {"T", "20"}, {"U", "21"},
		{"V", "22"}, {"Y", "25"}, {"Z", "26"},
	}
	for _, e := range entries {
		bt.Insert(e.k, []byte(e.v))
	}
	bt.Dump()
	fmt.Println("Traverse (in-order):")
	for _, kv := range bt.Traverse() {
		fmt.Printf("%s:%s ", kv.Key, string(kv.Value))
	}
	fmt.Println("\nSearch M:")
	if v, ok := bt.Search("M"); ok {
		fmt.Println("M ->", string(v))
	}
	fmt.Println("Delete M, X, A")
	bt.Delete("M")
	bt.Delete("X")
	bt.Delete("A")
	bt.Dump()
	fmt.Println("Traverse after deletes:")
	for _, kv := range bt.Traverse() {
		fmt.Printf("%s:%s ", kv.Key, string(kv.Value))
	}
	fmt.Println()
}

func TestBplus(t *testing.T) {
	bpt := NewBPlusTree(2)
	bpt.Insert(10, "a")
	bpt.Insert(20, "b")
	bpt.Insert(5, "c")
	bpt.Insert(11, "a")
	bpt.Insert(21, "b")
	bpt.Insert(6, "c")
	val, ok := bpt.Search(5)
	result := bpt.Traversal()
	for _, kv := range *result {
		fmt.Printf("%v:%s ", kv.key, kv.value.(string))
	}

	bpt.Delete(5)
	bpt.Delete(6)
	fmt.Println(val, ok) // Output: c true
}

func Test_avlTree(t *testing.T) {
	avlTree := &AVLTree{root: InitAVLNode(5)}

	vals := []int{4, 6, 2, 1, 8, 9, 11, 2, 32, 21, 15}

	//vals := []int{11}
	for _, val := range vals {
		avlTree.insert(val)
	}
	avlTree.delete(9)
	avlTree.delete(11)
	avlTree.delete(6)
	avlTree.delete(2)
	avlTree.delete(4)
	avlTree.insert(35)
	avlTree.delete(1)

	//avlTree.preorder()
	avlTree.inorder()
}

func Test_RBTree(t *testing.T) {
	mainRbt()
}
