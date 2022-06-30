package dotsync

import (
	"crypto/sha1"
	"errors"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"github.com/spf13/afero"
)

// Diffs two files by reading them in chunks of 1024 bytes
// first checks if the two files are of different size and if no
// difference is found continues to check the difference
// returns true if difference is found
const idxFileName = ".idx"

func openFile(filePath string) (afero.File, error) {
	fs := aferoFs.Fs
	idxFile, err := fs.OpenFile(filepath.Join(
		filePath, idxFileName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return idxFile, nil
}

// IndexFiles takes all files in the config and generates
// a sha1 index of the file. The index file is assumed to be
// in the same order as the config listed. It is considered
// corrupt if out of order and a new 
// TODO: Prob not return error from this
func (s *SyncConfig) IndexFiles(idxFile afero.File) error {
	for _, filePath := range s.Files {
		f, err := aferoFs.Open(filePath); if err != nil {
			return err
		}
		hash, err := sha1FileHash(f)
		if err != nil {
			return err
		}
		s.FileIndexes[filePath] = hash
	}
	return nil
}

func DiffFiles(filePath1 string, filePath2 string) (bool, error) {
	if filePath1 == "" || filePath2 == "" {
		return true, errors.New("Empty path supplied")
	}
	file1, err := aferoFs.Open(filePath1)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	file2, err := aferoFs.Open(filePath2)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	// Make sure we close files when we are done
	defer file1.Close()
	defer file2.Close()
	// First check if there is a difference in file size
	fileHash1, err := sha1FileHash(file1)
	if err != nil {
		return true, err
	}
	fileHash2, err := sha1FileHash(file2)
	if err != nil {
		return true, err
	}
	return bytes.Compare(fileHash1,fileHash2) != 0, nil
}

// Generates a SHA1 hash of the file
func sha1FileHash(file afero.File) ([]byte, error) {
	shaHasher := sha1.New()
	if _, err := io.Copy(shaHasher, file); err != nil {
		return nil, err
	}
	return shaHasher.Sum(nil), nil
}
