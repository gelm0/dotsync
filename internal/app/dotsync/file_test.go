package dotsync

import (
	"os"
	"testing"

	"crypto/rand"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	otherPath   = "/tmp/other"
	dotsyncPath = "/tmp/dotsync"
)

func openFiles(fileNames ...string) []afero.File {
	files := []afero.File{}
	for _, path := range fileNames {
		f, err := aferoFs.Open(path)
		if err != nil {
			panic(err)
		}
		files = append(files, f)
	}
	return files
}

func fillFileWithData(filePath string) {
	f, err := aferoFs.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	data := make([]byte, 1024)
	_, err = rand.Read(data)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
}

func initalise() ([]string, []string) {
	aferoFs.Fs = afero.NewMemMapFs()
	aferoFs.MkdirAll(otherPath, os.ModePerm)
	aferoFs.MkdirAll(dotsyncPath, os.ModePerm)
	currentFiles := []string{"1", "2", "3"}
	newFiles := []string{"4", "5", "6"}
	for i, _ := range currentFiles {
		fullSyncPath := filepath.Join(dotsyncPath, currentFiles[i])
		fullOtherPath := filepath.Join(otherPath, newFiles[i])
		fillFileWithData(fullSyncPath)
		fillFileWithData(fullOtherPath)
		currentFiles[i] = fullSyncPath
		newFiles[i] = fullOtherPath
	}
	return currentFiles, newFiles
}

func TestInitialiseIndex(t *testing.T) {
	_, newFiles := initalise()
	index := InitialiseIndex(newFiles)
	assert.Equal(t, len(index.New), 3)
	i := 0
	for _, v := range index.New {
		assert.Equal(t, v.Path, newFiles[i])
		assert.Equal(t, v.Perm, os.FileMode(0666))
		i += 1
	}
}

func TestParseNonExistingIndexFile(t *testing.T) {
	_, newFiles := initalise()
	index := InitialiseIndex(newFiles)
	index.ParseIndexFile(dotsyncPath)
	ok, err := aferoFs.Exists(filepath.Join(dotsyncPath, IndexFileName))
	if err != nil {
		panic(err)
	}
	assert.Equal(t, ok, true)
}

func TestParseIndexFile(t *testing.T) {

}
