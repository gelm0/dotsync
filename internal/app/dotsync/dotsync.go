package dotsync

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type GitConfig struct {
	URL     string `yaml:"url"`
	KeyFile string `yaml:"sshKey"`
	Branch  string `yaml:"branch,omitempty"`
	Remote  string `yaml:"remote,omitempty`
}

type SyncConfig struct {
	GitConfig GitConfig `yaml:"gitconfig"`
	Path      string    `yaml:"path"`
	Files     []string  `yaml:"files"`
}

const (
	DotSyncPath = ".dotsync"
)

// Errors
var (
	ErrMissingConfig     = errors.New("missing config")
	ErrMissingGitConfig  = errors.New("no git credentials supplied")
	ErrMissingGitURL     = errors.New("missing git url")
	ErrMissingSSHKeyFile = errors.New("missing sshkey file")
	ErrInvalidSSHKey     = errors.New("sshkey invalid")
)

var aferoFs = afero.Afero{
	Fs: afero.NewOsFs(),
}

var log *logrus.Logger

// Setup inital logging to stderr
func init() {
	log = logrus.New()
}

func SetupLogging(logDir string) *logrus.Logger {
	// TODO: Syslog instead?
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  filepath.Join(logDir, "info.log"),
		logrus.ErrorLevel: filepath.Join(logDir, "error.log"),
	}
	log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))
	return log
}

// 1. Files are synced from local to git repository
// 2. Files are synced from git to local

// Just nonempty validation for now
func (s *SyncConfig) Validate() error {
	if s.GitConfig == (GitConfig{}) {
		return ErrMissingGitConfig
	}

	config := &s.GitConfig
	if config.URL == "" {
		return ErrMissingGitURL
	}

	if config.Branch == "" {
		config.Branch = "main"
	}

	if config.Remote == "" {
		config.Remote = "origin"
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

func getConfigPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Error("Failed to get user home", err)
		return ""
	}
	return filepath.Join(homePath, DotSyncPath, "config")
}

// Utility function if config does not exists
// creates dotsync directory in user home and an empty config
func createConfig(configPath string) {
}

func OpenSyncConfig() (SyncConfig, error) {
	config := SyncConfig{}
	configPath := getConfigPath()
	if configPath == "" {
		return config, ErrMissingConfig
	}
	bytes, err := aferoFs.ReadFile(configPath)
	if err != nil  {
		if errors.Is(err, os.ErrNotExist) {
			createConfig(configPath)	
		} else {
			return config, err
		}
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
	syncConfig, err := OpenSyncConfig()
	if err != nil {
		log.WithField("path", getConfigPath()).Panic("Failed to open config file. Error: ", err)
		os.Exit(1)
	}
	err = syncConfig.Validate()
	if err != nil {
		log.Error("Failed to validate config", err)
		os.Exit(1)
	}
	// This feels weird and clunky, think of something better
	SetupLogging(syncConfig.Path)
	index := InitialiseIndex(syncConfig.Files)
	index.ParseIndexFile(syncConfig.Path)
	newIndex, err := index.CopyAndCleanup(syncConfig.Path)
	// TODO: Git operations
	if err != nil {
		log.Error("File indexing ran into an issue", err)
	}

	repository, err := NewRepository(syncConfig)
	if err != nil {
		log.Error("Failed to open repository", err)
	}

	err = repository.tryAndUpdate()
	if err != nil {
		log.Info("Failed to update repository")
		os.Exit(1)
	}
	// cleanup old files
	for k := range index.Current {
		fileToSync := filepath.Join(syncConfig.Path, k)
		repository.removeFile(fileToSync)
	}
	// Add new files
	for k := range newIndex {
		fileToSync := filepath.Join(syncConfig.Path, k)
		repository.addFile(fileToSync)
	}
	if (len(index.Current) > 0 || len(newIndex) > 0) {
		commitMessage := fmt.Sprintf("synced %d, removed %d files", len(newIndex), len(index.Current))
		repository.commit(commitMessage)
		repository.push()
		log.Info(commitMessage)
	} else {
		// To spammy?
		log.Info("No changes")
	}
}

// TODO later.
// Syncs the origin to local files. The files are kept in the default path
// specified /tmp/dotsync. Files can be either be specified to be kept in a certain
// folder such as /home/<username> or each file can be moved out to a specified
// path. If path is omitted from file, the files are kept in folder specified by user
// or if omitted in the default path /tmp/dotsync
func SyncLocal() {}
