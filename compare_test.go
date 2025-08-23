package tinypdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareBytes(t *testing.T) {
	tests := []struct {
		name    string
		data1   []byte
		data2   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "identical empty slices",
			data1:   []byte{},
			data2:   []byte{},
			wantErr: false,
		},
		{
			name:    "identical single bytes",
			data1:   []byte{42},
			data2:   []byte{42},
			wantErr: false,
		},
		{
			name:    "identical larger slices",
			data1:   []byte("hello world this is a test"),
			data2:   []byte("hello world this is a test"),
			wantErr: false,
		},
		{
			name:    "different single bytes",
			data1:   []byte{42},
			data2:   []byte{43},
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "different lengths - first shorter",
			data1:   []byte("hello"),
			data2:   []byte("hello world"),
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "different lengths - second shorter",
			data1:   []byte("hello world"),
			data2:   []byte("hello"),
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "difference at beginning",
			data1:   []byte("aello world"),
			data2:   []byte("hello world"),
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "difference at end",
			data1:   []byte("hello worla"),
			data2:   []byte("hello world"),
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "difference in middle",
			data1:   []byte("hello xorld"),
			data2:   []byte("hello world"),
			wantErr: true,
			errMsg:  "documents are different",
		},
		{
			name:    "large identical data",
			data1:   bytes.Repeat([]byte("test"), 1000),
			data2:   bytes.Repeat([]byte("test"), 1000),
			wantErr: false,
		},
		{
			name:    "large different data",
			data1:   bytes.Repeat([]byte("test"), 1000),
			data2:   bytes.Repeat([]byte("tent"), 1000),
			wantErr: true,
			errMsg:  "documents are different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompareBytes(tt.data1, tt.data2, false)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("CompareBytes() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Fatalf("CompareBytes() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("CompareBytes() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestComparePDFs(t *testing.T) {
	tests := []struct {
		name    string
		data1   []byte
		data2   []byte
		wantErr bool
	}{
		{
			name:    "identical data",
			data1:   []byte("PDF content here"),
			data2:   []byte("PDF content here"),
			wantErr: false,
		},
		{
			name:    "different data",
			data1:   []byte("PDF content here"),
			data2:   []byte("PDF content there"),
			wantErr: true,
		},
		{
			name:    "empty readers",
			data1:   []byte{},
			data2:   []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader1 := bytes.NewReader(tt.data1)
			reader2 := bytes.NewReader(tt.data2)

			err := ComparePDFs(reader1, reader2, false)

			if tt.wantErr && err == nil {
				t.Error("ComparePDFs() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ComparePDFs() unexpected error = %v", err)
			}
		})
	}
}

func TestComparePDFFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	file1Path := filepath.Join(tempDir, "test1.pdf")
	file2Path := filepath.Join(tempDir, "test2.pdf")
	file3Path := filepath.Join(tempDir, "test3.pdf")
	missingPath := filepath.Join(tempDir, "missing.pdf")

	// Write test data
	data1 := []byte("PDF test content 1")
	data2 := []byte("PDF test content 1") // identical to data1
	data3 := []byte("PDF test content 2") // different from data1

	if err := os.WriteFile(file1Path, data1, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2Path, data2, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file3Path, data3, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		file1   string
		file2   string
		wantErr bool
	}{
		{
			name:    "identical files",
			file1:   file1Path,
			file2:   file2Path,
			wantErr: false,
		},
		{
			name:    "different files",
			file1:   file1Path,
			file2:   file3Path,
			wantErr: true,
		},
		{
			name:    "missing second file (should succeed)",
			file1:   file1Path,
			file2:   missingPath,
			wantErr: false,
		},
		{
			name:    "missing first file",
			file1:   missingPath,
			file2:   file2Path,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePDFFiles(tt.file1, tt.file2, false)

			if tt.wantErr && err == nil {
				t.Error("ComparePDFFiles() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ComparePDFFiles() unexpected error = %v", err)
			}
		})
	}
}

// Test the specific bug we found in CompareBytes
func TestCompareBytesEdgeCases(t *testing.T) {
	// Test the bug where files of different sizes aren't properly detected
	t.Run("different sizes should be detected", func(t *testing.T) {
		data1 := []byte("short")
		data2 := []byte("much longer data")

		err := CompareBytes(data1, data2, false)
		if err == nil {
			t.Error("CompareBytes() should detect different file sizes")
		}
	})

	// Test the off-by-one error in the loop condition
	t.Run("single byte difference at end", func(t *testing.T) {
		data1 := []byte("a")
		data2 := []byte("b")

		err := CompareBytes(data1, data2, false)
		if err == nil {
			t.Error("CompareBytes() should detect single byte difference")
		}
	})

	// Test differences exactly at 16-byte boundaries
	t.Run("difference at 16-byte boundary", func(t *testing.T) {
		// Create 16 bytes that are identical, then differ
		data1 := make([]byte, 17)
		data2 := make([]byte, 17)

		// Make first 16 bytes identical
		for i := 0; i < 16; i++ {
			data1[i] = byte(i)
			data2[i] = byte(i)
		}

		// Make 17th byte different
		data1[16] = 1
		data2[16] = 2

		err := CompareBytes(data1, data2, false)
		if err == nil {
			t.Error("CompareBytes() should detect difference at 16-byte boundary")
		}
	})
}

// Test print diff functionality
func TestCompareBytesWithPrintDiff(t *testing.T) {
	// Capture stdout to verify diff is printed
	data1 := []byte("hello")
	data2 := []byte("world")

	// We can't easily capture stdout in a Unit test, but we can at least
	// verify the function doesn't panic when printDiff is true
	err := CompareBytes(data1, data2, true)
	if err == nil {
		t.Error("CompareBytes() should detect difference and print diff")
	}
}
