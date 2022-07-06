package dotsync

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)
/*
# Algorithm
We don't longer care about filepaths other then indexing them from that path
for now we don't care about history either we can implement that later if interesting
so

for all files in syncConfig
	generate sha1hashes

read current index file in dotsync path

if no index file
	sync all files
	generate index file
else
	keep set of sha1hash
	remove relative complement of sha1hash in indexfile
*/

// Each file index contains
// Path of original file
// Sha1hash
// Original filemode
// Any errors while trying to index it
type FileIndex struct {
	Path	string
	Hash	string
	Perm	os.FileMode
}

type Indexes struct {
	Current	map[FileIndex]bool
	New		map[FileIndex]bool
}


func InitialiseIndex(files []string) (index *Indexes) {
	index = &Indexes{
		Current: make(map[FileIndex]bool),
		New: make(map[FileIndex]bool),
	}
	for _, filePath := range files {
		if filePath == "" {
			continue
		}
		file, err := aferoFs.Open(filePath)
		if err != nil {
			log.WithField("file", filePath).
				Error(err)
			continue
		}
		defer file.Close()
		hash, err := sha1FileHash(file)

	}
	return
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

