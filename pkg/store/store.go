package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store manages file-based record storage rooted at a configured directory.
// Each Schema maps to its own file within the Store directory.
type Store struct {
	dir string
}

// New creates a Store rooted at dir, creating the directory if needed.
// dir must be non-empty; callers are responsible for providing the store dir.
func New(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("store dir must not be empty")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Dir returns the base directory of the store.
func (s *Store) Dir() string {
	return s.dir
}

// Schema[T] is a typed, file-backed collection of records acting as a table.
// Each Schema maps to a single .jsonl file in the Store directory.
// T is the record type and serves as the schema definition.
type Schema[T any] struct {
	path string
	mu   sync.RWMutex
}

// Open returns a Schema[T] for the given name within s.
// The name becomes the filename (name.jsonl) in the store directory.
func Open[T any](s *Store, name string) *Schema[T] {
	return &Schema[T]{
		path: filepath.Join(s.dir, name+".jsonl"),
	}
}

// ReadAll returns all records from the schema's file.
// Returns an empty slice (not an error) when the file does not yet exist.
func (sc *Schema[T]) ReadAll() ([]T, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	f, err := os.Open(sc.path)
	if os.IsNotExist(err) {
		return []T{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open schema file: %w", err)
	}
	defer f.Close()

	var records []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec T
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("decode record: %w", err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read schema file: %w", err)
	}
	return records, nil
}

// Write appends a single record to the schema's file.
// If the record implements Validator, it is validated before writing.
func (sc *Schema[T]) Write(record T) error {
	if v, ok := any(record).(Validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode record: %w", err)
	}

	f, err := os.OpenFile(sc.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open schema file for write: %w", err)
	}
	defer f.Close()

	if _, err = fmt.Fprintf(f, "%s\n", data); err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	return nil
}

// WriteAll replaces all records in the schema's file atomically.
// If any record implements Validator, all are validated before any write occurs.
func (sc *Schema[T]) WriteAll(records []T) error {
	for i, rec := range records {
		if v, ok := any(rec).(Validator); ok {
			if err := v.Validate(); err != nil {
				return fmt.Errorf("validation failed on record %d: %w", i, err)
			}
		}
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	tmp := sc.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	w := bufio.NewWriter(f)
	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("encode record: %w", err)
		}
		w.Write(data)
		w.WriteByte('\n')
	}
	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("flush: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmp, sc.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("commit write: %w", err)
	}
	return nil
}

// Validator is an optional interface a record type can implement
// to validate itself before being written to a Schema.
type Validator interface {
	Validate() error
}
