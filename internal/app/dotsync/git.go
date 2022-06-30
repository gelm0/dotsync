package dotsync

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/afero"
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

type PlainGitOperations interface {
	plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
	plainOpen(path string) (*git.Repository, error)
}

type gitExtension struct{}

func (g *gitExtension) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}

func (g *gitExtension) plainOpen(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

// Creates and returns a Repository. If nothing exist
// a clone operation will be performed
// Returns the repository and a nil error if sucessful
func NewRepository(s SyncConfig) (*Repository, error) {
	return newRepository(s, nil)
}

func newRepository(s SyncConfig, g PlainGitOperations) (*Repository, error) {
	if g == nil {
		g = &gitExtension{}
	}

	remoteURL := s.GitConfig.URL
	fs := aferoFs.Fs
	authObject, err := afero.ReadFile(fs, s.GitConfig.KeyFile)
	if err != nil {
		return nil, err
	}
	var repo = &git.Repository{}
	var r = &Repository{}
	branch := s.GitConfig.Branch
	if _, err := fs.Stat(filepath.Join(DotSyncPath, ".git")); errors.Is(err, os.ErrNotExist) {
		repo, err = cloneSSH(remoteURL, branch, []byte(authObject), g)
		if err != nil {
			return nil, err
		}
	} else {
		repo, err = g.plainOpen(DotSyncPath)
		if err != nil {
			return nil, err
		}
	}
	r.Repo = repo
	return r, nil
}

// Clones a repository using ssh url formatting and a valid sshKey read as byte slice
// Returns error if unable to clone the specified repository url
func cloneSSH(remoteURL, branch string, sshKey []byte, g PlainGitOperations) (*git.Repository, error) {
	publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
	if err != nil {
		return nil, err
	}
	r, err := g.plainClone(DotSyncPath, false, &git.CloneOptions{
		URL:        remoteURL,
		Progress:   os.Stdout,
		RemoteName: branch,
		Auth:       publicKey,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
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
		return errors.New("no remotename supplied")
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
	// No basic auth found, trying sshAuth
	if sshKey == nil {
		return ErrInvalidSSHKey
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

func (r *Repository) Add(filePaths []string) error {
	workTree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	for _, file := range filePaths {
		_, err := workTree.Add(file)
		if err != nil {
			// Bit naive implementation we should do something better
			// with error handling here
			return err
		}
	}
	return nil
}
