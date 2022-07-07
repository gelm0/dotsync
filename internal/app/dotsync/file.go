package dotsync

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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
type FileInfo struct {
	Path string
	Perm os.FileMode
	// If file was failed to index
	Failed bool
}

// Key is the hash of the file
type Indexes struct {
	Current map[string]FileInfo
	New     map[string]FileInfo
}

func InitialiseIndex(dotsyncPath string, files []string) (index *Indexes) {
	index = &Indexes{
		Current: make(map[string]FileInfo),
		New:     make(map[string]FileInfo),
	}
	for _, filePath := range files {
		failed := false
		if filePath == "" {
			continue
		}
		file, err := aferoFs.Open(filePath)
		if err != nil {
			failed = true
			log.WithField("file", filePath).
				Error("Failed to open file", err)
			continue
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			failed = true
			log.WithField("file", filePath).
				Error("Failed to stat", err)
			continue
		}
		hash, err := sha1FileHash(file)
		if err != nil {
			failed = true
			log.WithField("file", filePath).
				Error("Failed to create hash", err)
			continue
		}

		index.New[hash] = FileInfo{
			Path:   filePath,
			Perm:   fileInfo.Mode(),
			Failed: failed,
		}

	}
	index.ParseIndexFile(dotsyncPath)
	return
}

func (index *Indexes) ParseIndexFile(configPath string) {
	file, err := aferoFs.Open(filepath.Join(configPath, ".idx"))
	if err != nil {
		log.Debug("Failed to open index file. Creating new")
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var path, hash string
		var fileMode uint32
		fmt.Sscanf(scanner.Text(), "%s:%s:%d", &path, &hash, &fileMode)
		// TODO: Nil check values
		index.Current[hash] = FileInfo{
			Path:   path,
			Perm:   os.FileMode(fileMode),
			Failed: false,
		}
	}

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
	return fileHash1 != fileHash2, nil
}

// Generates a SHA1 hash of the file
func sha1FileHash(file afero.File) (string, error) {
	shaHasher := sha1.New()
	if _, err := io.Copy(shaHasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(shaHasher.Sum(nil)), nil
}
