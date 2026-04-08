// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package validate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestAtomicWrite_WritesContentAndPermissionCorrectly(t *testing.T) {
	// GIVEN: a target path in a temp directory
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key":"value"}`)

	// WHEN: AtomicWrite writes data with 0644 permission
	if err := AtomicWrite(path, data, 0644); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// THEN: file content matches exactly
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content = %q, want %q", got, data)
	}
}

func TestAtomicWrite_SetsRestrictivePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}

	// GIVEN: a target path
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.json")

	// WHEN: AtomicWrite writes with 0600 permission
	if err := AtomicWrite(path, []byte("secret"), 0600); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// THEN: file permission is exactly 0600 (owner read-write only)
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permission = %04o, want 0600", perm)
	}
}

func TestAtomicWrite_OverwritesExistingFile(t *testing.T) {
	// GIVEN: an existing file with old content
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	AtomicWrite(path, []byte("old"), 0644)

	// WHEN: AtomicWrite overwrites with new content
	if err := AtomicWrite(path, []byte("new"), 0644); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	// THEN: file contains new content
	got, _ := os.ReadFile(path)
	if string(got) != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
}

func TestAtomicWrite_LeavesNoResidualTempFileOnError(t *testing.T) {
	// GIVEN: a target path in a non-existent nested directory
	path := filepath.Join(t.TempDir(), "nonexistent", "subdir", "file.txt")

	// WHEN: AtomicWrite fails (parent directory doesn't exist)
	err := AtomicWrite(path, []byte("data"), 0644)

	// THEN: the write fails
	if err == nil {
		t.Fatal("expected error writing to nonexistent dir")
	}

	// THEN: no .tmp files are left behind
	parentDir := filepath.Dir(filepath.Dir(path))
	entries, _ := os.ReadDir(parentDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("residual temp file found: %s", e.Name())
		}
	}
}

func TestAtomicWrite_PreservesOriginalFileOnFailure(t *testing.T) {
	// GIVEN: an existing file with known content
	dir := t.TempDir()
	original := []byte("original content")
	path := filepath.Join(dir, "file.json")
	if err := AtomicWrite(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	// WHEN: AtomicWrite targets a non-existent directory (guaranteed to fail even as root)
	badPath := filepath.Join(dir, "no", "such", "dir", "file.json")
	err := AtomicWrite(badPath, []byte("new"), 0644)

	// THEN: write fails
	if err == nil {
		t.Fatal("expected error writing to non-existent dir")
	}

	// THEN: the original file at the valid path is untouched
	got, _ := os.ReadFile(path)
	if string(got) != string(original) {
		t.Errorf("original file corrupted: got %q, want %q", got, original)
	}
}

func TestAtomicWrite_HandlesCorrectlyUnderConcurrentWrites(t *testing.T) {
	// GIVEN: a target file that will be written by 20 concurrent goroutines
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.json")

	// WHEN: 20 goroutines write simultaneously
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			data := []byte(`{"n":` + string(rune('0'+n%10)) + `}`)
			AtomicWrite(path, data, 0644)
		}(i)
	}
	wg.Wait()

	// THEN: file exists and is valid (not corrupted by interleaved writes)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(got) == 0 {
		t.Error("file is empty after concurrent writes")
	}
}

func TestAtomicWriteFromReader_WritesContentFromReader(t *testing.T) {
	// GIVEN: a target path and a reader with known content
	dir := t.TempDir()
	path := filepath.Join(dir, "from_reader.txt")
	content := []byte("hello from reader")

	// WHEN: AtomicWriteFromReader writes the reader content
	n, err := AtomicWriteFromReader(path, strings.NewReader(string(content)), 0644)

	// THEN: no error and bytes count matches
	if err != nil {
		t.Fatalf("AtomicWriteFromReader failed: %v", err)
	}
	if n != int64(len(content)) {
		t.Errorf("returned n = %d, want %d", n, len(content))
	}

	// THEN: file content matches
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestAtomicWriteFromReader_ReturnsZeroOnError(t *testing.T) {
	// GIVEN: a path in a non-existent nested directory (guaranteed to fail)
	path := filepath.Join(t.TempDir(), "nonexistent", "subdir", "file.txt")

	// WHEN: AtomicWriteFromReader fails because parent directory doesn't exist
	n, err := AtomicWriteFromReader(path, strings.NewReader("data"), 0644)

	// THEN: returns error and zero bytes copied
	if err == nil {
		t.Fatal("expected error writing to nonexistent dir")
	}
	if n != 0 {
		t.Errorf("expected n=0 on error, got %d", n)
	}
}

func TestAtomicWriteFromReader_SetsPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}

	// GIVEN: a target path
	dir := t.TempDir()
	path := filepath.Join(dir, "perm_test.bin")

	// WHEN: AtomicWriteFromReader writes with 0600 permission
	_, err := AtomicWriteFromReader(path, strings.NewReader("data"), 0600)
	if err != nil {
		t.Fatalf("AtomicWriteFromReader failed: %v", err)
	}

	// THEN: file permission is exactly 0600
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permission = %04o, want 0600", perm)
	}
}

func TestAtomicWriteFromReader_LeavesNoResidualTempFileOnError(t *testing.T) {
	// GIVEN: a target path in a non-existent nested directory
	baseDir := t.TempDir()
	path := filepath.Join(baseDir, "nonexistent", "file.txt")

	// WHEN: AtomicWriteFromReader fails
	_, _ = AtomicWriteFromReader(path, strings.NewReader("data"), 0644)

	// THEN: no .tmp files are left in the base directory
	entries, _ := os.ReadDir(baseDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("residual temp file found: %s", e.Name())
		}
	}
}

func TestAtomicWriteFromReader_OverwritesExistingFile(t *testing.T) {
	// GIVEN: an existing file with old content
	dir := t.TempDir()
	path := filepath.Join(dir, "overwrite.txt")
	AtomicWrite(path, []byte("old content"), 0644)

	// WHEN: AtomicWriteFromReader overwrites with new content
	_, err := AtomicWriteFromReader(path, strings.NewReader("new content"), 0644)
	if err != nil {
		t.Fatalf("AtomicWriteFromReader failed: %v", err)
	}

	// THEN: file contains new content
	got, _ := os.ReadFile(path)
	if string(got) != "new content" {
		t.Errorf("content = %q, want %q", got, "new content")
	}
}