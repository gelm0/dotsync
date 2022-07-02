package dotsync

import (
	"testing"

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

func TestDiffFilesIdentical(t *testing.T) {
	files := openFiles(vimrc, vimrcIdentical)
	response, err := DiffFiles(files[0], files[1])
	assert.Equal(t, err, nil)
	assert.Equal(t, response, false)
}

func TestDiffFilesEmptyFilesIdentical(t *testing.T) {
	files := openFiles(emptyFile, emptyFile)
	response, err := DiffFiles(files[0], files[1])
	assert.Equal(t, err, nil)
	assert.Equal(t, response, false)
}

func TestDiffFilesSameSizeDifference(t *testing.T) {
	files := openFiles(vimrc, vimrcSameSizeDiff)
	response, err := DiffFiles(files[0], files[1])
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}

func TestDiffFilesNewlineIntroduced(t *testing.T) {
	files := openFiles(vimrc, vimrcDiff1)
	response, err := DiffFiles(files[0], files[1])
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}
func TestDiffFilesChangesIntroduced(t *testing.T) {
	files := openFiles(vimrc, vimrcDiff2)
	response, err := DiffFiles(files[0], files[1])
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}

//func TestDiffFilesFileNotExist(t *testing.T) {
//	files := openFiles(vimrc, nonExistingFile)
//	response, err := DiffFiles(files[0], files[1])
//	assert.Error(t, err)
//	assert.Equal(t, response, true)
//}
