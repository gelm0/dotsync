package dotsync

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/afero"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	Repo *git.Repository
	Auth *ssh.PublicKeys
	Remote string
	Branch string
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

	auth, err := ssh.NewPublicKeys("git", []byte(authObject), "")
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(filepath.Join(DotSyncPath, ".git")); errors.Is(err, os.ErrNotExist) {
		repo, err = cloneSSH(remoteURL, branch, auth, g)
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
	r.Auth = auth
	r.Branch = s.GitConfig.Branch
	r.Remote = s.GitConfig.Remote
	return r, nil
}

// Clones a repository using ssh url formatting and a valid sshKey read as byte slice
// Returns error if unable to clone the specified repository url
func cloneSSH(remoteURL, branch string, auth *ssh.PublicKeys , g PlainGitOperations) (*git.Repository, error) {
	r, err := g.plainClone(DotSyncPath, false, &git.CloneOptions{
		URL:        	remoteURL,
		Progress:  		os.Stdout,
		ReferenceName: 	plumbing.NewBranchReferenceName(branch),
		Auth:       	auth,
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

func (r *Repository) Fetch() error {
	err := r.Repo.Fetch(&git.FetchOptions{
		RemoteName: r.Remote,
		Auth: r.Auth,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func (r *Repository) Pull() error {
	w, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{
		RemoteName: r.Remote,
		Auth: r.Auth,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

// Pushes current commited files to remote. Assumes HTTPs or SSH depending on auth method
// that is supplied to the function
func (r *Repository) Push() error {
	return r.Repo.Push(&git.PushOptions{
		RemoteName: r.Remote,
		Auth:       r.Auth,
	})
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

// Tries and update the repository with a git pull. Tries and reset the repository to the the HEAD of origin
// returns an error if that fails
func (r *Repository) TryAndUpdate() error {
	err := r.Fetch(remoteName)
	if err != nil {
		log.WithField("remoteName", remoteName).Error(err)
		return err
	}

	remoteRef, err := r.Repo.Reference(plumbing.ReferenceName("refs/remotes/"+ remoteName +"/"+r.Branch), true)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Remote": r.Remote,
			"Branch": r.Branch,
		}).Error(err)
		return err
	}
	localRef, err := repo.Reference(plumbing.ReferenceName("HEAD"), true)
	if err != nil {
		rlog.Errorf("Failed to get local reference for HEAD: %v", err)
		return
	}

	return nil
}
