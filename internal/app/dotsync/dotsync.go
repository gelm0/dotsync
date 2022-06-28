package dotsync

import (
	// 	"errors"

	// 	"strings"
	// 	"time"
	"errors"
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
	URL     string   `yaml:"url,omitempty"`
	KeyFile string   `yaml:"sshKey,omitempty"`
	Branch  string   `yaml:"branch,omitempty"`
}

type SyncConfig struct {
	GitConfig GitConfig `yaml:"gitconfig"`
	Files   []string 	`yaml:"files"`
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

// 1. Files are synced from local to git repository
// 2. Files are synced from git to local

func ValidateSyncConfig() error {
	return nil
}

func OpenSyncConfig(configPath string, afero afero.Afero) (SyncConfig, error) {
	bytes, err := afero.ReadFile(configPath)
	config := SyncConfig{}

	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
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