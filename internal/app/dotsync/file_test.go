package dotsync

import (
	"math/rand"
	"os"
	"testing"

	"crypt/rand"
	"path/filepath"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

/* TODO: Make this generated data */
const rootFolder = "../../../"
const (
	vimrc             = rootFolder + "test/data/vimrc"
	vimrcDiff1        = rootFolder + "test/data/vimrc_newline"
	vimrcDiff2        = rootFolder + "test/data/vimrc_new_options"
	vimrcSameSizeDiff = rootFolder + "test/data/vimrc_diff_size"
	vimrcIdentical    = rootFolder + "test/data/vimrc_identical"
	emptyFile         = rootFolder + "test/data/emptyfile"
	nonExistingFile   = rootFolder + "test/data/dont_exist"
)

const (
	otherPath = "/tmp/other"
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
	f, err := aferoFs.OpenFile(filePath, os.O_CREATE | os.O_RDWR, 0666)
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

func initalise() {
	aferoFs.Fs = afero.NewMemMapFs()
	aferoFs.MkdirAll(otherPath, os.ModePerm)
	aferoFs.MkdirAll(dotsyncPath, os.ModePerm)
	syncFiles := []string{"1", "2", "3"}
	otherFiles := []string{"4","5", "6"}
	for i, _ := range syncFiles {
		fillFileWithData(filepath.Join(dotsyncPath, syncFiles[i]))
		fillFileWithData(filepath.Join(otherPath, otherFiles[i]))
	}
}

