package tree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Simple key-value pair
type KV struct {
	Key   string
	Value []byte // nil with Tombstone==true indicates deletion
}

// On-disk format per entry: [keyLen:int32][key:bytes][valLen:int32][val:bytes]
// If valLen == -1 => tombstone (deletion)

// SSTable holds metadata for a single file
type SSTable struct {
	Path  string
	Keys  []string         // in-memory index of keys (sorted)
	Index map[string]int64 // key -> file offset
	mutex sync.RWMutex
}

// DB is a minimal LSM DB
type DB struct {
	dir         string
	memtable    map[string]*KV
	memSize     int // approximate bytes in memtable
	wal         *os.File
	sstables    []*SSTable // newest first
	mutex       sync.RWMutex
	flushThresh int
	//compacting   bool
	compactionMu sync.Mutex
}

// Open opens/creates a DB in directory dir
func Open(dir string, flushThreshold int) (*DB, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	walPath := filepath.Join(dir, "wal.log")
	wal, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	db := &DB{
		dir:         dir,
		memtable:    make(map[string]*KV),
		memSize:     0,
		wal:         wal,
		sstables:    nil,
		flushThresh: flushThreshold,
	}

	// Load existing SSTables from dir (newest sorted first)
	if err := db.loadSSTables(); err != nil {
		return nil, err
	}
	// Replay WAL to populate memtable
	if err := db.replayWAL(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if db.wal != nil {
		_ = db.wal.Sync()
		_ = db.wal.Close()
		db.wal = nil
	}
	return nil
}

// write WAL entry: "P|keyLen|key|valLen|val\n" or "D|keyLen|key\n"
// For simplicity we use binary append consistent with SST format.
func (db *DB) appendWAL(kv *KV) error {
	var buf bytes.Buffer
	if kv.Value == nil {
		// tombstone
		if err := buf.WriteByte('D'); err != nil {
			return err
		}
		if err := writeString(&buf, kv.Key); err != nil {
			return err
		}
	} else {
		if err := buf.WriteByte('P'); err != nil {
			return err
		}
		if err := writeString(&buf, kv.Key); err != nil {
			return err
		}
		if err := writeBytes(&buf, kv.Value); err != nil {
			return err
		}
	}
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if _, err := db.wal.Write(buf.Bytes()); err != nil {
		return err
	}
	// flush to disk for durability
	if err := db.wal.Sync(); err != nil {
		return err
	}
	return nil
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}
func writeBytes(w io.Writer, b []byte) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}
func readString(r io.Reader) (string, error) {
	var ln int32
	if err := binary.Read(r, binary.BigEndian, &ln); err != nil {
		return "", err
	}
	if ln < 0 {
		return "", errors.New("negative length")
	}
	buf := make([]byte, ln)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
func readBytes(r io.Reader) ([]byte, error) {
	var ln int32
	if err := binary.Read(r, binary.BigEndian, &ln); err != nil {
		return nil, err
	}
	if ln < 0 {
		// convention: -1 tombstone handled by caller
		return nil, nil
	}
	buf := make([]byte, ln)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Put inserts or updates a key
func (db *DB) Put(key string, value []byte) error {
	if key == "" {
		return errors.New("empty key")
	}
	kv := &KV{Key: key, Value: value}
	if err := db.appendWAL(kv); err != nil {
		return err
	}
	db.mutex.Lock()
	db.memtable[key] = kv
	db.memSize += len(key) + len(value)
	shouldFlush := db.memSize >= db.flushThresh
	db.mutex.Unlock()
	if shouldFlush {
		go db.Flush() // background flush
	}
	return nil
}

// Delete inserts a tombstone
func (db *DB) Delete(key string) error {
	kv := &KV{Key: key, Value: nil} // nil indicates tombstone
	if err := db.appendWAL(kv); err != nil {
		return err
	}
	db.mutex.Lock()
	// We store tombstone entry pointer
	db.memtable[key] = kv
	db.memSize += len(key) + 1
	db.mutex.Unlock()
	if db.memSize >= db.flushThresh {
		go db.Flush()
	}
	return nil
}

// Get searches memtable first, then SSTables (newest-first)
func (db *DB) Get(key string) ([]byte, bool, error) {
	// memtable
	db.mutex.RLock()
	if kv, ok := db.memtable[key]; ok {
		db.mutex.RUnlock()
		if kv.Value == nil {
			return nil, false, nil // tombstone
		}
		return kv.Value, true, nil
	}
	db.mutex.RUnlock()

	// search SSTables newest to oldest
	db.mutex.RLock()
	ssts := append([]*SSTable(nil), db.sstables...) // copy
	db.mutex.RUnlock()

	for _, s := range ssts {
		if v, found, err := s.Get(key); err == nil && found {
			if v == nil { // tombstone
				return nil, false, nil
			}
			return v, true, nil
		} else if err != nil {
			return nil, false, err
		}
	}
	return nil, false, nil
}

// List returns key-values in [start, end) lexicographically. If end=="" then to end.
func (db *DB) List(start, end string) ([]KV, error) {
	// Collect iterators from memtable and sstables.
	db.mutex.RLock()
	mem := make([]KV, 0, len(db.memtable))
	for _, kv := range db.memtable {
		mem = append(mem, *kv)
	}
	ssts := append([]*SSTable(nil), db.sstables...)
	db.mutex.RUnlock()

	// Build merged map: latest wins
	tmp := make(map[string]*KV)
	for i := len(ssts) - 1; i >= 0; i-- { // oldest -> newest so newest override later
		entries, err := ssts[i].All()
		if err != nil {
			return nil, err
		}
		for _, kv := range entries {
			// If not set yet, set (older entries)
			if _, ok := tmp[kv.Key]; !ok {
				cpy := kv
				tmp[kv.Key] = &cpy
			}
		}
	}
	// Now memtable overrides everything
	for _, kv := range mem {
		cpy := kv
		tmp[kv.Key] = &cpy
	}

	// Collect keys in range and sort
	keys := make([]string, 0, len(tmp))
	for k := range tmp {
		if (start == "" || k >= start) && (end == "" || k < end) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := make([]KV, 0, len(keys))
	for _, k := range keys {
		v := tmp[k]
		if v.Value == nil {
			// tombstone => skip
			continue
		}
		out = append(out, *v)
	}
	return out, nil
}

// Flush writes memtable to an SSTable and clears WAL + memtable
func (db *DB) Flush() error {
	// prevent multiple simultaneous flushes
	db.mutex.Lock()
	if len(db.memtable) == 0 {
		db.mutex.Unlock()
		return nil
	}

	// Snapshot memtable
	snap := make([]KV, 0, len(db.memtable))
	for _, kv := range db.memtable {
		snap = append(snap, *kv)
	}
	// reset memtable (we'll write snapshot to SSTable)
	db.memtable = make(map[string]*KV)
	db.memSize = 0
	// rotate wal: close and recreate
	if db.wal != nil {
		_ = db.wal.Close()
	}
	walPath := filepath.Join(db.dir, "wal.log")
	newWal, err := os.OpenFile(walPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		db.mutex.Unlock()
		return err
	}
	db.wal = newWal
	db.mutex.Unlock()

	// Create SSTable from snap
	if err := db.writeSSTable(snap); err != nil {
		return err
	}
	return nil
}

func (db *DB) writeSSTable(entries []KV) error {
	if len(entries) == 0 {
		return nil
	}
	// Sort entries by key
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })

	// File path
	ts := time.Now().UnixNano()
	filename := fmt.Sprintf("sst_%d.sst", ts)
	path := filepath.Join(db.dir, filename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write entries sequentially and build index
	index := make(map[string]int64)
	for _, e := range entries {
		off, _ := f.Seek(0, io.SeekCurrent)
		index[e.Key] = off

		// key
		if err := writeString(f, e.Key); err != nil {
			return err
		}
		// value
		if e.Value == nil {
			// tombstone -> write -1
			if err := binary.Write(f, binary.BigEndian, int32(-1)); err != nil {
				return err
			}
		} else {
			if err := writeBytes(f, e.Value); err != nil {
				return err
			}
		}
	}

	// build SSTable in-memory index
	keys := make([]string, 0, len(index))
	for k := range index {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sst := &SSTable{
		Path:  path,
		Keys:  keys,
		Index: index,
	}
	// Prepend as newest
	db.mutex.Lock()
	db.sstables = append([]*SSTable{sst}, db.sstables...)
	db.mutex.Unlock()

	return nil
}

// loadSSTables loads existing sstable filenames from dir
func (db *DB) loadSSTables() error {
	files, err := filepath.Glob(filepath.Join(db.dir, "sst_*.sst"))
	if err != nil {
		return err
	}
	// sort newest-first by filename (timestamp in name)
	sort.Slice(files, func(i, j int) bool {
		// descending
		return files[i] > files[j]
	})
	ssts := make([]*SSTable, 0, len(files))
	for _, p := range files {
		s, err := loadSSTable(p)
		if err != nil {
			return err
		}
		ssts = append(ssts, s)
	}
	db.sstables = ssts
	return nil
}

// replayWAL reads wal.log and populates memtable
func (db *DB) replayWAL() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	walPath := filepath.Join(db.dir, "wal.log")
	f, err := os.Open(walPath)
	if err != nil {
		// no wal yet
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	for {
		var t [1]byte
		_, err := f.Read(t[:])
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		tag := t[0]
		if tag == 'D' {
			key, err := readString(f)
			if err != nil {
				return err
			}
			db.memtable[key] = &KV{Key: key, Value: nil}
		} else if tag == 'P' {
			key, err := readString(f)
			if err != nil {
				return err
			}
			val, err := readBytes(f)
			if err != nil {
				return err
			}
			db.memtable[key] = &KV{Key: key, Value: val}
		} else {
			return fmt.Errorf("unknown wal tag: %c", tag)
		}
	}
	return nil
}

// loadSSTable builds in-memory index for a file (scans keys)
func loadSSTable(path string) (*SSTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	index := make(map[string]int64)
	keys := make([]string, 0)
	for {
		off, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		// read key
		var ln int32
		if err := binary.Read(f, binary.BigEndian, &ln); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}
		if ln < 0 {
			return nil, fmt.Errorf("bad key len")
		}
		kbuf := make([]byte, ln)
		if _, err := io.ReadFull(f, kbuf); err != nil {
			return nil, err
		}
		key := string(kbuf)

		// read val len
		var vln int32
		if err := binary.Read(f, binary.BigEndian, &vln); err != nil {
			return nil, err
		}
		// skip val
		if vln >= 0 {
			if _, err := f.Seek(int64(vln), io.SeekCurrent); err != nil {
				return nil, err
			}
		}
		index[key] = off
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return &SSTable{Path: path, Keys: keys, Index: index}, nil
}

// Get key from an SSTable (reads value at offset)
func (s *SSTable) Get(key string) ([]byte, bool, error) {
	s.mutex.RLock()
	off, ok := s.Index[key]
	s.mutex.RUnlock()
	if !ok {
		return nil, false, nil
	}
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return nil, false, err
	}
	// read key (skip)
	if _, err := readString(f); err != nil {
		return nil, false, err
	}
	// read val
	// We need to peek valLen
	var vln int32
	if err := binary.Read(f, binary.BigEndian, &vln); err != nil {
		return nil, false, err
	}
	if vln == -1 {
		return nil, true, nil // tombstone
	}
	if vln < 0 {
		return nil, false, fmt.Errorf("bad val len")
	}
	buf := make([]byte, vln)
	if _, err := io.ReadFull(f, buf); err != nil {
		return nil, false, err
	}
	return buf, true, nil
}

// All returns all key-value pairs from sstable (in order). Tombstones included with nil Value.
func (s *SSTable) All() ([]KV, error) {
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := make([]KV, 0)
	for {
		key, err := readString(f)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}
		// read val
		var vln int32
		if err := binary.Read(f, binary.BigEndian, &vln); err != nil {
			return nil, err
		}
		if vln == -1 {
			out = append(out, KV{Key: key, Value: nil})
			continue
		}
		buf := make([]byte, vln)
		if _, err := io.ReadFull(f, buf); err != nil {
			return nil, err
		}
		out = append(out, KV{Key: key, Value: buf})
	}
	return out, nil
}

// Compact merges all sstables into one (simple single-level compaction)
func (db *DB) Compact() error {
	// Simple lock to avoid concurrent compactions
	db.compactionMu.Lock()
	defer db.compactionMu.Unlock()

	db.mutex.RLock()
	ssts := append([]*SSTable(nil), db.sstables...) // newest first
	db.mutex.RUnlock()
	if len(ssts) <= 1 {
		return nil
	}

	// Merge from oldest->newest so newest overrides older
	merged := make(map[string]*KV)
	for i := len(ssts) - 1; i >= 0; i-- {
		entries, err := ssts[i].All()
		if err != nil {
			return err
		}
		for _, kv := range entries {
			if _, ok := merged[kv.Key]; !ok {
				cpy := kv
				merged[kv.Key] = &cpy
			}
		}
	}
	// Also apply memtable (if any) - but typically compaction runs on disk tables only
	db.mutex.RLock()
	for _, kv := range db.memtable {
		cpy := *kv
		merged[kv.Key] = &cpy
	}
	db.mutex.RUnlock()

	// Create sorted list of KV for new SSTable (drop tombstones)
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	entries := make([]KV, 0, len(keys))
	for _, k := range keys {
		kv := merged[k]
		// If tombstone, we may drop it to reclaim space
		if kv.Value == nil {
			// skip tombstone (garbage collect)
			continue
		}
		entries = append(entries, *kv)
	}

	// Write new sstable
	if err := db.writeSSTable(entries); err != nil {
		return err
	}

	// Remove old sstable files
	db.mutex.Lock()
	old := db.sstables
	// keep only the newly created one (first element)
	if len(db.sstables) > 0 {
		db.sstables = db.sstables[:1]
	}
	db.mutex.Unlock()
	for _, s := range old {
		// don't delete new one
		_ = os.Remove(s.Path)
	}
	return nil
}
