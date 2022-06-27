package dotsync

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

/* TODO: Make this generated data */
const (
	vimrc             = "test/data/vimrc"
	vimrcDiff1        = "test/data/vimrc_newline"
	vimrcDiff2        = "test/data/vimrc_new_options"
	vimrcSameSizeDiff = "test/data/vimrc_diff_size"
	vimrcIdentical    = "test/data/vimrc_identical"
	emptyFile         = "test/data/emptyfile"
	nonExistingFile   = "test/data/dont_exist"
)

func TestDiffFilesIdentical(t *testing.T) {
	response, err := DiffFiles(vimrc, vimrcIdentical)
	assert.Equal(t, err, nil)
	assert.Equal(t, response, false)
}

func TestDiffFilesEmptyFilesIdentical(t *testing.T) {
	response, err := DiffFiles(emptyFile, emptyFile)
	assert.Equal(t, err, nil)
	assert.Equal(t, response, false)
}

func TestDiffFilesSameSizeDifference(t *testing.T) {
	response, err := DiffFiles(vimrc, vimrcSameSizeDiff)
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}

func TestDiffFilesNewlineIntroduced(t *testing.T) {
	response, err := DiffFiles(vimrc, vimrcDiff1)
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}
func TestDiffFilesChangesIntroduced(t *testing.T) {
	response, err := DiffFiles(vimrc, vimrcDiff2)
	assert.Equal(t, err, nil)
	assert.Equal(t, response, true)
}

func TestDiffFilesFileNotExist(t *testing.T) {
	response, err := DiffFiles(vimrc, nonExistingFile)
	assert.Error(t, err)
	assert.Equal(t, response, true)
}
