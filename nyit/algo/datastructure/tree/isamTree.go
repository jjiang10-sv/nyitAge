package tree

import "sort"

type Record struct {
	Key   int
	Value interface{}
}

type Block struct {
	Entries  []Record
	Overflow *Block
}

type ISAM struct {
	Index     []int
	Primary   []*Block
	BlockSize int
}

func NewISAM(records []Record, blockSize int) *ISAM {
	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})

	var index []int
	var primary []*Block

	for i := 0; i < len(records); i += blockSize {
		end := i + blockSize
		if end > len(records) {
			end = len(records)
		}
		block := &Block{Entries: records[i:end]}
		primary = append(primary, block)
		index = append(index, block.Entries[len(block.Entries)-1].Key)
	}
	return &ISAM{Index: index, Primary: primary, BlockSize: blockSize}
}

func (isam *ISAM) FindBlock(key int) *Block {
	i := sort.Search(len(isam.Index), func(i int) bool { return isam.Index[i] >= key })
	if i < len(isam.Index) {
		return isam.Primary[i]
	}
	return isam.Primary[len(isam.Primary)-1]
}

// ---------------- INSERT -------------------

func (isam *ISAM) Insert(key int, value interface{}) {
	block := isam.FindBlock(key)

	// insert into primary if space
	if len(block.Entries) < isam.BlockSize {
		i := sort.Search(len(block.Entries), func(i int) bool { return block.Entries[i].Key >= key })
		block.Entries = append(block.Entries[:i],
			append([]Record{{key, value}}, block.Entries[i:]...)...)
		return
	}

	// otherwise insert into overflow chain
	current := block
	for {
		if current.Overflow == nil {
			current.Overflow = &Block{Entries: []Record{{key, value}}}
			return
		}
		if len(current.Overflow.Entries) < isam.BlockSize {
			current.Overflow.Entries = append(current.Overflow.Entries, Record{key, value})
			return
		}
		current = current.Overflow
	}
}

// ---------------- SEARCH -------------------

func (isam *ISAM) Search(key int) (interface{}, bool) {
	block := isam.FindBlock(key)

	i := sort.Search(len(block.Entries), func(i int) bool { return block.Entries[i].Key >= key })
	if i < len(block.Entries) && block.Entries[i].Key == key {
		return block.Entries[i].Value, true
	}

	for current := block.Overflow; current != nil; current = current.Overflow {
		for _, entry := range current.Entries {
			if entry.Key == key {
				return entry.Value, true
			}
		}
	}
	return nil, false
}

// ---------------- DELETE -------------------

func deleteFromBlock(block *Block, key int) bool {
	for i, entry := range block.Entries {
		if entry.Key == key {
			block.Entries = append(block.Entries[:i], block.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func (isam *ISAM) Delete(key int) bool {
	block := isam.FindBlock(key)

	// 1. try primary block
	if deleteFromBlock(block, key) {
		return true
	}

	// 2. try overflow chain
	prev := block
	current := block.Overflow
	for current != nil {
		if deleteFromBlock(current, key) {
			// if overflow block becomes empty, drop it
			if len(current.Entries) == 0 {
				prev.Overflow = current.Overflow
			}
			return true
		}
		prev = current
		current = current.Overflow
	}
	return false
}

// ---------------- TRAVERSE -------------------

func (isam *ISAM) Traverse() []Record {
	var result []Record

	for _, block := range isam.Primary {
		// primary entries in sorted order
		result = append(result, block.Entries...)

		// overflow blocks are not sorted
		for current := block.Overflow; current != nil; current = current.Overflow {
			result = append(result, current.Entries...)
		}
	}
	return result
}
