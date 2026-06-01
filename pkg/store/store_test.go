package store_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/danlafeir/cli-go/pkg/store"
)

type note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type validatedNote struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func (v validatedNote) Validate() error {
	if v.Text == "" {
		return fmt.Errorf("text must not be empty")
	}
	return nil
}

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestNew_EmptyDir(t *testing.T) {
	_, err := store.New("")
	if err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestNew_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/nested/store"
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.Dir() != dir {
		t.Fatalf("Dir() = %q, want %q", s.Dir(), dir)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("dir not created: %v", err)
	}
}

func TestReadAll_EmptyWhenNoFile(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[note](s, "notes")
	records, err := sc.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestWrite_AppendAndReadAll(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[note](s, "notes")

	for i := 1; i <= 3; i++ {
		if err := sc.Write(note{ID: i, Text: fmt.Sprintf("note %d", i)}); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	records, err := sc.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	for i, r := range records {
		if r.ID != i+1 {
			t.Errorf("record %d: ID = %d, want %d", i, r.ID, i+1)
		}
	}
}

func TestWriteAll_ReplacesRecords(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[note](s, "notes")

	sc.Write(note{ID: 1, Text: "old"})
	sc.Write(note{ID: 2, Text: "old"})

	if err := sc.WriteAll([]note{{ID: 10, Text: "new"}}); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	records, err := sc.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 1 || records[0].ID != 10 {
		t.Fatalf("expected [{10 new}], got %v", records)
	}
}

func TestWriteAll_Empty(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[note](s, "notes")

	sc.Write(note{ID: 1})
	if err := sc.WriteAll([]note{}); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	records, _ := sc.ReadAll()
	if len(records) != 0 {
		t.Fatalf("expected 0 records after WriteAll([]), got %d", len(records))
	}
}

func TestValidator_BlocksInvalidRecord(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[validatedNote](s, "vnotes")

	err := sc.Write(validatedNote{ID: 1, Text: ""})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	err = sc.Write(validatedNote{ID: 1, Text: "ok"})
	if err != nil {
		t.Fatalf("unexpected error for valid record: %v", err)
	}
}

func TestValidator_WriteAll_ValidatesAll(t *testing.T) {
	s := newTestStore(t)
	sc := store.Open[validatedNote](s, "vnotes")

	err := sc.WriteAll([]validatedNote{{ID: 1, Text: "ok"}, {ID: 2, Text: ""}})
	if err == nil {
		t.Fatal("expected validation error from WriteAll, got nil")
	}
}

func TestIndependentSchemas_SeparateFiles(t *testing.T) {
	s := newTestStore(t)
	notes := store.Open[note](s, "notes")
	other := store.Open[note](s, "other")

	notes.Write(note{ID: 1})
	notes.Write(note{ID: 2})
	other.Write(note{ID: 99})

	nr, _ := notes.ReadAll()
	or_, _ := other.ReadAll()

	if len(nr) != 2 {
		t.Errorf("notes: expected 2 records, got %d", len(nr))
	}
	if len(or_) != 1 || or_[0].ID != 99 {
		t.Errorf("other: expected [{99}], got %v", or_)
	}
}
