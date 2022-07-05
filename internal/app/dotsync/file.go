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

// Internal struct to keep tab of which files we had issues with while trying to determine
// if they were sync
type SyncedFile struct {
	FilePath 	string
	Error		error
}


// Test all files if they have been changed
// Returns a set of files that need to be resynced
// TODO: I don't know if the struct SyncedFiles is actually necessary.
// Using it as a precaution as I want to display some meaningful error message
// in the gui while syncing files, but it might be that it should be used further
// in the process and we don't actually care about it right now
func (s *SyncConfig) IndexFiles() (resync []SyncedFile) {
	for _, localFile := range s.Files {
		if localFile == "" {
			continue
		}
		fileName := filepath.Base(localFile) 
		originFile := filepath.Join(s.Path, fileName)
		file1, errLocal := aferoFs.Open(localFile)
		if errLocal != nil {
				resync = append(resync, SyncedFile{
					FilePath: localFile,
					Error: errLocal,
				})
				log.WithField("file", localFile).
					Error(errLocal)
				continue
			}
		file2, errOrigin := aferoFs.Open(originFile)
		if errOrigin != nil {
				resync = append(resync, SyncedFile{
					FilePath: localFile,
					Error: errLocal,
				})
				log.WithField("file", localFile).
					Error(errOrigin)
				continue
			}
		defer file1.Close()
		defer file2.Close()
		ok, err := DiffFiles(file1, file2)
		if err != nil {
			log.WithFields(logrus.Fields{
				"file1": file1.Name(),
				"file2": file2.Name(),
			}).Error(err)
		}
		if !ok {
			resync = append(resync, SyncedFile{
				FilePath: localFile,
				Error: err,
			})
		}

	}
	return
}

// Internal struct to keep track on what we visited in WalkDir
type walked struct {
	filesToRemove []string
	files map[string]bool
}

// TODO: WE only support a flat file structure, this means that we can have no duplicate files
func (w *walked) isUnwatched(path string, info os.DirEntry, err error) error {
	if info.IsDir() {
		return nil
	}
	if _, ok := w.files[info.Name()]; !ok {
		w.filesToRemove = append(w.filesToRemove, info.Name())
	}
	return nil
}

// Returns all the files that we no longer wish to watch
// i.e. sync
func (s *SyncConfig) FindUnwatchedFiles() (unWatched []string) {
	localPath := s.Path
	w := walked{
		filesToRemove: []string{},
		files: make(map[string]bool),
	}
	// Initialise map
	for _, f := range s.Files {
		//Strip the local file path
		fileName := filepath.Base(f)
		w.files[fileName] = true
	}
	err := filepath.WalkDir(localPath, w.isUnwatched)
	if err != nil {
		log.Error(err)
	}
	return w.filesToRemove
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
