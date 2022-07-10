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
// Original filemode
// Any errors while trying to index it
type FileInfo struct {
	Path string
	Perm os.FileMode
	// If file was failed to index
	Failed bool
}

// Contains needed fileinfo for new files inexes as well
// as the previous files parsed from index file
// Key is the hash of the file
type Indexes struct {
	Current map[string]FileInfo
	New     map[string]FileInfo
}

const (
	IndexFileName = ".idx"
)

// Returns an Indexes struct with the current index of tracked files
// as well as the previous tracked parsed from the index file
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
	file, err := aferoFs.Open(filepath.Join(configPath, IndexFileName))
	if err != nil {
		log.Debug("Failed to open index file. Creating new")
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var path, hash string
		var fileMode uint32
		n, err := fmt.Sscanf(scanner.Text(), "%s:%s:%d", &path, &hash, &fileMode)
		if err != nil {
			log.Error("Problem when scanning indexfile ", err)
		}
		// TODO: Nil check values
		if n == 3 {
			index.Current[hash] = FileInfo{
				Path:   path,
				Perm:   os.FileMode(fileMode),
				Failed: false,
			}
		} else {
			if hash != "" {
				index.Current[hash] = FileInfo{
					Path:   path,
					Perm:   os.FileMode(fileMode),
					Failed: true,
				}
			}
			log.WithFields(logrus.Fields{
				"Number": n,
				"Path":   path,
				"Perm":   os.FileMode(fileMode),
			}).Warning("Missing one or more fields in indexfile")
		}
	}
}

// Creates indexfile if not exist otherwise truncates and
// writes the new index
func writeIndexFile(configPath string, files map[string]FileInfo) error {
	filePath := filepath.Join(configPath, IndexFileName)
	file, err := aferoFs.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	for k, v := range files {
		line := fmt.Sprintf("%s:%s:%d", k, v.Path, v.Perm)
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFiles(files map[string]FileInfo) error {

}

func cleanupOldFiles(files map[string]FileInfo) error {

}

func (index *Indexes) CopyAndCleanup(configPath string) {
	// Current all the files we want to keep
	// Old all the files that we want to get rid of
	// Diff these and create a list over what we need to copy
	// and what we should remove
	copy := make(map[string]FileInfo)
	newIndex := make(map[string]FileInfo)
	for k, v := range index.New {
		if _, ok := index.Current[k]; !ok {
			copy[k] = v
		} else {
			delete(index.Current, k)
		}
		newIndex[k] = v
	}
	cleanup := &index.Current
	err := cleanupOldFiles(cleanup)
	if err != nil {
		log.WithField("File", filePath).Error(err)
	}
	err = copyFiles(copy)
	if err != nil {
		log.WithField("File", filePath).Error(err)
	}
	err = writeIndexFile(newIndex)
	if err != nil {
		log.WithField("File", filePath).Error(err)
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
