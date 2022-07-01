package dotsync

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/afero"
)

const idxFileName = ".idx"

func (s *SyncConfig) openIndexFile() (afero.File, error) {
	fs := aferoFs.Fs
	idxFile, err := fs.OpenFile(filepath.Join(
		s.Path, idxFileName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return idxFile, nil
}

// Save this for later ideas
func (s *SyncConfig) IndexFile(filePath string, idxFile afero.File) ([]byte, error) {
	hash := []byte{}
	f, err := aferoFs.Open(filePath); if err != nil {
		return hash, err
	}
	hash, err = sha1FileHash(f)
	if err != nil {
		return hash, err
	}
	return hash, nil
}

// Test all files if they have been changed
// Returns a set of files that need to be resynced
func (s *SyncConfig) IndexFiles() ([]string, error) {
	resyncFiles := []string{}
	for _, localFile := range s.Files {
		if localFile == "" {
			continue
		}
		splits := strings.Split(localFile, "/")
		fileName := splits[len(splits) - 1]
		originFile := filepath.Join(s.Path, fileName)
		file1, err := aferoFs.Open(localFile)
		if err != nil {
			// Might change this to log directly later, for now we return it
			return resyncFiles, err
		}
		file2, err := aferoFs.Open(originFile)
		if err != nil {
			// Might change this to log directly later, for now we return it
			return resyncFiles, err
		}
		defer file1.Close()
		defer file2.Close()
		ok, err := DiffFiles(file1, file2)
		if err != nil {
			return resyncFiles, err
		}
		if !ok {
			resyncFiles = append(resyncFiles, localFile)
		}

	}
	return resyncFiles, nil
}

func DiffFiles(file1 afero.File, file2 afero.File) (bool, error) {
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
