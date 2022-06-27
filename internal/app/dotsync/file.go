package dotsync

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// Diffs two files by reading them in chunks of 1024 bytes
// first checks if the two files are of different size and if no
// difference is found continues to check the difference
// returns true if difference is found
func DiffFiles(filePath1 string, filePath2 string) (bool, error) {
	if filePath1 == "" || filePath2 == "" {
		return true, errors.New("Empty path supplied")
	}
	file1, err := os.Open(filePath1)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	file2, err := os.Open(filePath2)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	// Make sure we close files when we are done
	defer file1.Close()
	defer file2.Close()
	// First check if there is a difference in file size
	fileHash1, err := getFileHash(file1)
	if err != nil {
		return true, err
	}
	fileHash2, err := getFileHash(file2)
	if err != nil {
		return true, err
	}
	return fileHash1 != fileHash2, nil
}

// Generates a SHA256 hash of the file
func getFileHash(file *os.File) (string, error) {
	shaHasher := sha256.New()
	if _, err := io.Copy(shaHasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(shaHasher.Sum(nil)), nil
}
