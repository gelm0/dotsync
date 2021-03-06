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
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type repository struct {
	Repo   *git.Repository
	Auth   *ssh.PublicKeys
	Remote string
	Branch string
}

type gitOperations interface {
	Commit(commitMessage string) error
	Pull(remoteName string) error
	Push(remoteName, basicAuth string, sshKey []byte) error
	Add(filePaths []string) error
}

type plainGitOperations interface {
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
func NewRepository(s SyncConfig) (*repository, error) {
	return newRepository(s, nil)
}

func newRepository(s SyncConfig, g plainGitOperations) (*repository, error) {
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
	var r = &repository{}
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
func cloneSSH(remoteURL, branch string, auth *ssh.PublicKeys, g plainGitOperations) (*git.Repository, error) {
	r, err := g.plainClone(DotSyncPath, false, &git.CloneOptions{
		URL:           remoteURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Auth:          auth,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *repository) commit(commitMessage string) error {
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

func (r *repository) fetch() error {
	err := r.Repo.Fetch(&git.FetchOptions{
		RemoteName: r.Remote,
		Auth:       r.Auth,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func (r *repository) pull() error {
	w, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{
		RemoteName: r.Remote,
		Auth:       r.Auth,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

// Pushes current commited files to remote. Assumes HTTPs or SSH depending on auth method
// that is supplied to the function
func (r *repository) push() error {
	return r.Repo.Push(&git.PushOptions{
		RemoteName: r.Remote,
		Auth:       r.Auth,
	})
}

func (r *repository) addFile(filePath string) error {
	workTree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	_, err = workTree.Add(filePath)
	if err != nil {
		// Bit naive implementation we should do something better
		// with error handling here
		return err
	}
	return nil
}

func (r *repository) removeFile(filePath string) error {
	workTree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	_, err = workTree.Remove(filePath)
	if err != nil {
		// Bit naive implementation we should do something better
		// with error handling here
		return err
	}
	return nil
}

func (r *repository) add(filePaths []string) error {
	for _, file := range filePaths {
		err := r.addFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// Tries and update the repository with a git pull. Tries and reset the repository to the the HEAD of origin
// returns an error if that fails
func (r *repository) tryAndUpdate() error {
	err := r.pull()
	if err == nil {
		return nil
	}
	// Repo can't be updated, reset and retry

	remoteRef, err := r.Repo.Reference(
		plumbing.NewRemoteReferenceName(r.Remote, r.Branch), true)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Remote": r.Remote,
			"Branch": r.Branch,
		}).Error("Failed to fetch remote reference", err)
		return err
	}

	w, err := r.Repo.Worktree()
	if err != nil {
		log.Error("Failed to get worktree when trying to reset the repository")
		return err
	}

	err = w.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: remoteRef.Hash(),
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"Mode":   "Hard",
			"Commit": remoteRef.Hash(),
		}).Error("Failed to reset worktree")
		return err
	}
	// Confirm changes
	localRef, err := r.Repo.Reference(plumbing.HEAD, true)
	if err != nil {
		log.Error("Failed to fetch local reference", err)
		return err
	}

	if localRef != remoteRef {
		return errors.New("Failed to update")
	}

	return nil
}
