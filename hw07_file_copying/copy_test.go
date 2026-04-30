package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func verifyFileContent(t *testing.T, resultPath, expectedPath string) {
	t.Helper()
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}

	result, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}

	if string(result) != string(expected) {
		t.Errorf("file content does not match expected")
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name         string
		from         string
		offset       int64
		limit        int64
		expectedErr  error
		expectedFile string // Эталонный файл для проверки содержимого
	}{
		{
			name:         "copy entire file",
			from:         "testdata/input.txt",
			offset:       0,
			limit:        0,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset0_limit0.txt",
		},
		{
			name:         "copy with limit",
			from:         "testdata/input.txt",
			offset:       0,
			limit:        10,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset0_limit10.txt",
		},
		{
			name:         "copy with offset",
			from:         "testdata/input.txt",
			offset:       100,
			limit:        0,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset100_limit0.txt",
		},
		{
			name:         "copy with limit 1000",
			from:         "testdata/input.txt",
			offset:       0,
			limit:        1000,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset0_limit1000.txt",
		},
		{
			name:         "copy with limit 10000",
			from:         "testdata/input.txt",
			offset:       0,
			limit:        10000,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset0_limit10000.txt",
		},
		{
			name:         "copy with offset and limit",
			from:         "testdata/input.txt",
			offset:       100,
			limit:        1000,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset100_limit1000.txt",
		},
		{
			name:         "copy with offset 6000 and limit 1000",
			from:         "testdata/input.txt",
			offset:       6000,
			limit:        1000,
			expectedErr:  nil,
			expectedFile: "testdata/out_offset6000_limit1000.txt",
		},
		{
			name:         "offset exceeds file size",
			from:         "testdata/input.txt",
			offset:       1000000,
			limit:        0,
			expectedErr:  ErrOffsetExceedsFileSize,
			expectedFile: "",
		},
		{
			name:         "source file not found",
			from:         "testdata/nonexistent.txt",
			offset:       0,
			limit:        0,
			expectedErr:  os.ErrNotExist,
			expectedFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем временную директорию для выходного файла
			tmpDir := t.TempDir()
			toPath := filepath.Join(tmpDir, "output.txt")

			err := Copy(tt.from, toPath, tt.offset, tt.limit)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Проверяем, что файл создан
			if _, err := os.Stat(toPath); os.IsNotExist(err) {
				t.Errorf("output file was not created")
			}

			// Проверяем содержимое, если указан эталонный файл
			if tt.expectedFile != "" {
				verifyFileContent(t, toPath, tt.expectedFile)
			}
		})
	}
}
