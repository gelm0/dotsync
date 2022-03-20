package dotsync

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Repository struct {
	Repo *git.Repository
}

type GitOperations interface {
	Commit(commitMessage string) error
	Pull(remoteName string) error
	Push(remoteName, basicAuth string, sshKey []byte) error
	Add(filePaths []string) error
}

// Creates and returns a Repository. If nothing exist
// a clone operation will be performed
// Returns the repository and a nil error if sucessful
func NewRepository(s SyncConfig) (*Repository, error) {
	remoteURL := s.getURL()
	authObject, err := os.ReadFile(s.Credentials)
	if err != nil {
		return nil, err
	}
	branch := s.Branch
	if _, err := os.Stat(filepath.Join(RepoPath, ".git")); os.IsNotExist(err) {
		if s.HTTPS != "" {
			err := cloneHTTPS(remoteURL, branch, string(authObject))
			if err != nil {
				return nil, err
			}
		} else {
			err := cloneSSH(remoteURL, branch, []byte(authObject))
			if err != nil {
				return nil, err
			}
		}
	}
	r := &Repository{}
	repo, err := git.PlainOpen(RepoPath)
	if err != nil {
		return nil, err
	}
	r.Repo = repo
	return r, nil
}

func cloneHTTPS(remoteURL string, branch string, basicAuth string) error {
	username, password, err := splitBasicAuth(basicAuth)
	if err != nil {
		return err
	}
	_, err = git.PlainClone(RepoPath, false, &git.CloneOptions{
		URL:        remoteURL,
		RemoteName: branch,
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func splitBasicAuth(basicAuth string) (string, string, error) {
	credentialsSlice := strings.SplitN(basicAuth, ":", 1)
	if len(credentialsSlice) != 2 {
		return "", "", ErrInvalidBasicAuth
	}
	return credentialsSlice[0], credentialsSlice[1], nil
}

// Clones a repository using ssh url formatting and a valid sshKey read as byte slice
// Returns error if unable to clone the specified repository url
func cloneSSH(remoteURL, branch string, sshKey []byte) error {
	publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
	if err != nil {
		return err
	}
	_, err = git.PlainClone(RepoPath, false, &git.CloneOptions{
		URL:        remoteURL,
		Progress:   os.Stdout,
		RemoteName: branch,
		Auth:       publicKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Commit(commitMessage string) error {
	worktree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	// May have to check status of worktree here
	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name: "dotsync",
			When: time.Now(),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.Repo.CommitObject(commit)
	return err
}

func (r *Repository) Pull(remoteName string) error {
	if remoteName == "" {
		return errors.New("No remotename supplied")
	}
	w, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Pushes current commited files to remote. Assumes HTTPs or SSH depending on auth method
// that is supplied to the function
func (r *Repository) Push(remoteName, basicAuth string, sshKey []byte) error {
	if basicAuth != "" {
		username, password, err := splitBasicAuth(basicAuth)
		if err != nil {
			return err
		}
		err = r.Repo.Push(&git.PushOptions{
			RemoteName: remoteName,
			Auth: &http.BasicAuth{
				Username: username,
				Password: password,
			},
		})
		if err != nil {
			return err
		}
		return nil
	}
	// No basic auth found, trying sshAuth
	if sshKey == nil {
		return ErrNoCredentialsSupplied
	}
	publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
	if err != nil {
		return err
	}
	err = r.Repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth:       publicKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func Add(filePaths []string) error {
	return nil
}
