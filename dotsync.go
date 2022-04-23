package dotsync

import (
	// 	"errors"
	// 	"io/ioutil"
	// 	"path/filepath"
	// 	"strings"
	// 	"time"
	"errors"
	//
	// 	"github.com/go-git/go-git/v5"
	// 	"github.com/go-git/go-git/v5/plumbing/object"
	// 	"github.com/go-git/go-git/v5/plumbing/transport/http"
	// 	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	// 	"github.com/thanhpk/randstr" // Random string package
	// 	"gopkg.in/yaml.v2"
)

type SyncConfig struct {
	URL     string   `yaml: "url,omitempty"`
	KeyFile string   `yaml:"sshKey,omitempty"`
	Branch  string   `yaml:"branch,omitempty"`
	Files   []string `yaml:"files"`
}

const (
	RepoPath = "/tmp/dotsync"
)

// Errors
var (
	ErrNoCredentialsFile     = errors.New("missing credentials file")
	ErrInvalidBasicAuth      = errors.New("can not extract username, password")
	ErrNoCredentialsSupplied = errors.New("no basicauth or sshkey supplied")
	ErrNoRemoteURL           = errors.New("missing remote url")
)

// func copyFile(source string, destination string) error {
// 	input, err := ioutil.ReadFile(source)
// 	if err != nil {
// 		return err
// 	}
// 	// Make file permissions modifiable by input
// 	err = ioutil.WriteFile(destination, input, 0644)
// 	if err != nil {
// 		return err
// 	}
// }
//
// // Returns a new config
// // Looks in users homefolder by default
// func NewSyncConfig(path string) (*SyncConfig, error) {
// 	var configPath string
// 	if path == "" {
// 		homeDir, err := os.UserHomeDir()
// 		if err != nil {
// 			return nil, err
// 		}
// 		// Default path is assumed to be ~/.dotsync/dotsync.yaml
// 		configPath = filepath.Join(homeDir, ".dotsync", "dotsync.yaml")
// 	} else {
// 		configPath = path
// 	}
// 	// This might be unecessary, but who doesn't love descriptive errors
// 	_, err := os.Stat(configPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Try to unmarshal it
// 	s := SyncConfig{}
// 	os.ReadFile(configPath)
// 	err = yaml.Unmarshal([]byte(config), &s)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Validate the syncconfig
// 	err = s.ValidateAndSetDefaults()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &s, nil
// }
//
// // Checks if there is a difference between local and remote files that are
// // being watched
// // Returns a slice of files that are not in sync with remote
// func (r *Repository) DiffRemoteWithLocal() (filesNotInSync []string, err error) {
// 	err := r.repo.Pull()
// 	if err != nil {
// 		return err
// 	}
// 	for _, file := range s.Files {
// 		_, fileName := filepath.Split(file)
// 		remoteFile := filepath.Join(repoPath, fileName)
// 		difference, err := DiffFiles(file, remoteFile)
// 		if err != nil {
// 			// We ignore this and just log the error
// 			// TODO:
// 			// On second hand an error here could mean that
// 			// the file cannot be found on of the locations
// 			// which means that it should be synced
// 		}
// 		if difference {
// 			filesNotInSync = append(filesNotInSync, file)
// 		}
// 	}
// 	return
// }
//
// // Retrieves the remote name tied to the remoteUrl
// // If a remote name cannot be found an error will be returned
// func (r *Repository) GetRemoteName(remoteUrl string) (remoteName string, err error) {
// 	remotes, err := r.repo.Remotes()
// 	if err != nil {
// 		remoteName = ""
// 		return
// 	}
// 	for _, remote := range remotes {
// 		name := remote.Config.Name
// 		URLs := remote.Config().URLs
// 		for _, URL := range URLs {
// 			if URL == remoteURL {
// 				// Return the first remote that matches
// 				err = nil
// 				remoteName = name
// 				return
// 			}
// 		}
// 	}
// 	err = git.ErrRemoteNotFound
// 	remoteName = ""
// 	return
// }
//
// // Generates a random remote name of length 8
// // Returns remotename, error.
// func (r *Repository) CreateRemote() (remoteName string, err error) { // Refactor dependencies
// 	remoteName = randstr.Hex(8)
// 	remoteURL := s.getURL()
// 	r, err := r.repo.CreateRemote(&config.RemoteConfig{
// 		Name: remoteName,
// 		URLs: []string{remoteURL},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return
// }
//
//
//
//
// // Validates dotsyncs configuration and sets default values if not available
// // returns an descriptive error if a the config can not be validated
// func (s *SyncConfig) ValidateAndSetDefaults() error {
// 	// Check that a git url has been supplied
// 	if s.HTTPS == nil && s.SSH == nil {
// 		return ErrNoRemoteURL
// 	}
// 	// Check that we don't both have ssh and https url
// 	if s.HTTPS == nil && s.SSH == nil {
// 		return errors.New("Multiple git urls")
// 	}
// 	// Check that we have credentials if HTTPS is supplied
// 	if s.HTTPS != nil && s.Credentials == nil {
// 		return errors.New("Missing path to credentials file")
// 	}
// 	// Check that either credentials points to a valid file or that
// 	// the default ~/.ssh/id_rsa exists
// 	if s.SSH != nil {
// 		if s.Credentials != nil {
// 			if _, err := os.Stat(s.Credentials); os.IsNotExist(err) {
// 				return err
// 			}
// 		} else {
// 			homeDir, err := os.UserHomeDir()
// 			if err != nil {
// 				return err
// 			}
// 			defaultSSHKey := filepath.Join(homeDir, ".ssh", "id_rsa")
// 			if _, err := os.Stat(defaultSSHKey); os.IsNotExist(err) {
// 				return err
// 			}
// 			s.Credentials = defaultSSHKey
// 		}
// 	}
// 	// Branch name defaults to main if nothing else is set
// 	if s.Branch == nil {
// 		s.Branch = "main"
// 	}
// 	// Check that we atleast have one file that we can watch
// 	// otherwise no point of the program continuing after a validate
// 	if s.Files == nil || len(s.Files) == 0 {
// 		return errors.New("No files to watch")
// 	}
// }

// If remote is considered to be the single point of truth this function collects
// all files which have been changed locally and resets them to the state of the remote
// files
// Input a slice of files which are out of sync
// Returns an error if unable to sync the local files with remote
// func(r *Repository) SyncLocal(remoteName string, filesNotInSync []string) (error){
//     err := r.Pull(remoteName)
//     if err != nil {
//         return err
//     }
//     for _, file := range filesNotInSync {
//         _, fileName := filepath.Split(file)
//         remoteFile := filePath.Join(RepoPath, fileName)
//         err := copyFile(remoteFile, file); err != nil {
//             // Log error, must likely related to file permission
//         }
//     }
//     return nil
// }

// If local is considered the single point of truth this function collects all
// files which are have been changed and syncs it with remote.
// Input a slice of files which are out of sync
// Returns an error if unable to sync remote
// func(r *Repository) SyncRemote(remoteName string, basicAuth string, sshKey []byte filesNotInSync []string) (error){
//     err := r.Pull(remoteName)
//     if err != nil {
//         return err
//     }
//     workTree, err := r.repo.WorkTree()
//     if err != nil {
//         return err
//     }
//     // Probably convert this into a channel so we can run this async
//     filesAdded := make([]string, 0)
//     for _, file: range filesNotInSync {
//         _, fileName := filepath.Split(file)
//         remoteFile := filePath.Join(RepoPath, fileName)
//         if err := copyFile(file, remoteFile); err == nil {
//             err := workTree.Add(fileName)
//             if err == nil {
//                 filesAdded = append(filesAdded, fileName)
//             } else {
//                 // Log error
//             }
//         } else {
//             // Log error must likely related to file permission
//         }
//     }
//     // Construct commit message
//     commitMessage := "dotsync automated commit message. Updated the following files: "
//     for i, fileName := range filesAdded {
//         if i < len(filesAdded) -1 {
//             commitMessage += fileName + ", "
//         } else {
//             commitMessage += fileName
//         }
//     }
//     err := r.Commit(commitMessage)
//     if err != nil {
//         return err
//     }
//     r.Push(remoteName, basicAuth, sshKey)
// }
