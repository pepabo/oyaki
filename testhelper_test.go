package main

import (
	"os"
	"testing"
)

// readTestFile reads a test file and fatals on error.
func readTestFile(t testing.TB, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readTestFile: failed to read %s: %v", path, err)
	}
	return data
}

// isValidJPEG returns true when data begins with the JPEG SOI marker FF D8.
func isValidJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

// isValidWebP returns true when data has the RIFF....WEBP signature.
func isValidWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}
