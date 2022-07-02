package dotsync

import (
	// 	"errors"

	// 	"strings"
	// 	"time"
	"fmt"
	"errors"
	"os"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	//
	// 	"github.com/go-git/go-git/v5"
	// 	"github.com/go-git/go-git/v5/plumbing/object"
	// 	"github.com/go-git/go-git/v5/plumbing/transport/http"
	// 	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	// 	"github.com/thanhpk/randstr" // Random string package
	// 	"gopkg.in/yaml.v2"
)

type GitConfig struct {
	URL     string   `yaml:"url"`
	KeyFile	string   `yaml:"sshKey"`
	Branch  string   `yaml:"branch,omitempty"`
}

type SyncConfig struct {
	GitConfig 	GitConfig 	`yaml:"gitconfig"`
	Path		string		`yaml:"path"`
	Files		[]string 	`yaml:"files"`
}

const (
	DotSyncPath = "/tmp/dotsync"
)

// Errors
var (
	ErrMissingGitConfig 	= errors.New("no git credentials supplied")
	ErrMissingGitURL     	= errors.New("missing git url")
	ErrMissingSSHKeyFile 	= errors.New("missing sshkey file")
	ErrInvalidSSHKey		= errors.New("sshkey invalid")
)

var aferoFs = afero.Afero{
	Fs: afero.NewOsFs(),
}

// 1. Files are synced from local to git repository
// 2. Files are synced from git to local

// Just nonempty validation for now
func (s SyncConfig) Validate() error {
	if s.GitConfig == (GitConfig{}) {
		return ErrMissingGitConfig
	}

	config := s.GitConfig
	if config.URL == "" {
		return ErrMissingGitURL
	}

	if config.KeyFile == "" {
		return ErrInvalidSSHKey
	}

	if _, err := aferoFs.Stat(config.KeyFile); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", err, config.KeyFile)
	}

	if s.Path == "" {
		s.Path = DotSyncPath
	}
	
	return nil
}

func OpenSyncConfig(configPath string) (SyncConfig, error) {
	bytes, err := aferoFs.ReadFile(configPath)
	config := SyncConfig{}

	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}
	if err = config.Validate(); err != nil {
		return config, err
	}
	return config, nil
}

// Syncs the specified local files to a git repository
func SyncOrigin() {


}

// TODO later.
// Syncs the origin to local files. The files are kept in the default path
// specified /tmp/dotsync. Files can be either be specified to be kept in a certain
// folder such as /home/<username> or each file can be moved out to a specified
// path. If path is omitted from file, the files are kept in folder specified by user
// or if omitted in the default path /tmp/dotsync
func SyncLocal() {

}