package dotsync

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	diff "github.com/sergi/go-diff/diffmatchpatch"
    "github.com/thanhpk/randstr" // Random string package
	"gopkg.in/yaml.v2"
)

type SyncConfig struct {
	HTTPS       string      `yaml:"https,omitempty"`
	SSH         string      `yaml:"ssh,omitempty"`
	Credentials string      `yaml:"credentials,omitempty"`
	Branch      string      `yaml:"branch,omitempty"`
	Files       []string    `yaml:"files"`
}

type Repository struct {
    repo *git.Repository
}

type GitOperations interface {
    Commit(commitMessage string) error
    Push(remoteName, basicAuth string, sshKey []byte) error
    Pull(remoteName string) error
    Add(filePaths[]string) error
}

const (
    RepoPath = "/tmp/dotsync"
    ErrNoCredentialsFile = errors.new("Missing credentials file")
    ErrInvalidBasicAuth = errors.New("Can't extract username, password")
    ErrNoCredentialsSupplied = errors.new("No basicauth or sshkey supplied")
    ErrNoRemoteURL = errors.new("Missing remote url")
)

func copyFile(source string, destination string) (error) {
    input, err := ioutil.ReadFile(source)
    if err != nil {
        return err
    }
    // Make file permissions modifiable by input
    err := ioutil.WriteFile(destination, input, 0644)
    if err != nil {
        return err
    }
}

// Start by diffing first 1024 and last 1024 bytes
func diffChars(file1, file2 *os.File) (bool, error) {
    const bufferSize = 1024
    buffer1 := make([]byte, bufferSize)
    buffer2 := make([]byte, bufferSize)
    diffMatchPatch := diff.New()
    // We don't want to break the loop until the diff has been run
    breakLoop := false
    for {
        read1, err := file1.Read(buffer1)
        if err != nil {
            if err != io.EOF {
                return true, err }
            breakLoop = true
        }
        read2, err := file2.Read(buffer2)
        if err != nil {
            if err != io.EOF {
                return true, err
            }
            breakLoop = true
        }
        difference := diffMatchPatch.DiffMain(string(buffer1[:read1]), string(buffer2[:read2]), false)
        if len(difference) > 0 {
            for i := 0; i < len(difference); i++ {
                // 0 is equivalent to Equals, hence we check that everything that is not equal
                // is returned as a difference
                if difference[i].Type != 0 {
                    return true, nil
                }
            }
        }
        if breakLoop {
            break
        }
    }
    return false, nil
}

func diffSize(file1, file2 *os.File) (bool, error) {
    f1Info, err := file1.Stat()
    if err != nil {
        return true, err
    }
    f2Info, err := file1.Stat()
    if err != nil {
        return true, err
    }
    return f1Info.Size() != f2Info.Size(), nil
}

// Diffs two files by reading them in chunks of 1024 bytes
// first checks if the two files are of different size and if no
// difference is found continues to check the difference
// returns true if difference is found
func DiffFiles(filePath1 string, filePath2 string) (bool, error) {
    if filePath1 == "" || filePath2 == "" {
        return nil, errors.New("Empty path supplied")
    }
    file1, err := os.Open(filePath1)
    if err != nil {
        // Might change this to log directly later, for now we return it
        return true, err
    }
    file2, err := os.Open(filePath2)
    if err != nil {
        // Might change this to log directly later, for now we return it
        return true, err
    }
    // Make sure we close files when we are done
    defer file1.Close()
    defer file2.Close()
    // First check if there is a difference in file size
    sizeDiffers, err := diffSize(file1, file2)
    if sizeDiffers || err != nil {
        return sizeDiffers, err
    }
    // We didn't find any file size difference but we can't be sure that nothing changed
    return diffChars(file1, file2)
}

// Returns a new config
// Looks in users homefolder by default
func NewSyncConfig(path string) (*SyncConfig, error) {
    var configPath string
    if path == "" {
        homeDir, err := os.UserHomeDir()
        err != nil {
            return nil, err
        }
        // Default path is assumed to be ~/.dotsync/dotsync.yaml
        configPath = filepath.Join(homeDir, ".dotsync", "dotsync.yaml")
    } else {
        configPath = path
    }
    // This might be unecessary, but who doesn't love descriptive errors 
    _, err := os.Stat(configPath)
    err != nil {
        return nil, err
    }
    // Try to unmarshal it
    s := SyncConfig{}
    os.ReadFile(configPath)
    err := yaml.Unmarshal([]byte(config), &s)
    if err != nil {
        return nil, err
    }
    // Validate the syncconfig
    err := s.ValidateAndSetDefaults()
    if err != nil {
        return nil, err
    }
    return &s, nil
}

// Creates and returns a Repository. If nothing exist
// a clone operation will be performed
// Returns the repository and a nil error if sucessful
func NewRepository(s SyncConfig) (*Repository, error) {
    var remoteURL := s.getURL()
    authObject, err := os.ReadFile(s.Credentials)
    if err != nil {
        return nil, err
    }
    branch := s.Branch
    if _, err := os.Stat(filepath.Join(RepoPath, ".git"); os.IsNotExist(err) {
        if s.HTTPS != nil {
            err := CloneHTTPS(remoteUrl, string(authObject), branch)
        } else {
            err := CloneSSH(remoteUrl, []byte(authObject), branch)
        }
        if err != nil {
            return nil, err
        }
    }
    r := &Repository{}
    repo, err := git.PlainOpen(RepoPath)
    if err != nil {
        return nil, err
    }
    r.r = repo
    return r, nil
}

func(r *Repository) Pull(remoteName string) (error) {
    if remoteName == "" {
        return error.New("No remotename supplied")
    }
    w, err := r.repo.WorkTree()
    if err != nil {
        return err
    }
    err := w.Pull(&git.PullOptions{RemoteName: remoteName})
    if err != nil {
        return err
    }
    return nil
}

// Checks if there is a difference between local and remote files that are
// being watched
// Returns a slice of files that are not in sync with remote
func(r *Repository) DiffRemoteWithLocal() (filesNotInSync []string, error) {
    err := r.repo.Pull()
    if err != nil {
        return err
    }
    for _, file := range s.Files {
        _, fileName := filepath.Split(file)
        remoteFile := filepath.Join(repoPath, fileName)
        difference, err := DiffFiles(file, remoteFile)
        if err != nil {
            // We ignore this and just log the error
            // TODO:
            // On second hand an error here could mean that
            // the file cannot be found on of the locations
            // which means that it should be synced
        }
        if difference {
            filesNotInSync = append(filesNotInSync, file)
        }
    }
    return
}

// Pushes current commited files to remote. Assumes HTTPs or SSH depending on auth method
// that is supplied to the function
func(r *Repository) Push(remoteName, basicAuth: string, sshKey: []byte) error {
    if basicAuth != nil {
        username, password, err := splitBasicAuth(basicAuth)
        if err != nil {
            return err
        }
        err := r.repo.Push(&git.PushOptions{
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
    err := r.Repo.Push(&git.PushOptions{
        RemoteName: remoteName,
        Auth: publicKey,
    })
    if err != nil {
        return err
    }
    return nil
}

func(r *Repository) Commit(commitMessage string) {
   worktree, err := r.repo.Worktree()
   if err != nil {
       return err
   }
   // May have to check status of worktree here
   commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
      Author: &object.Signature{
        Name: "dotsync",
        Time: time.Now(),
      },
   })
   if err != nil {
       return err
   }
   _, err := r.repo.CommitObject(commit)
   return err
}

// Retrieves the remote name tied to the remoteUrl
// If a remote name cannot be found an error will be returned
func(r *Repository) GetRemoteName(remoteUrl string) (remoteName string, error) { //Refactor dependencies interface
    remotes, err := r.repo.Remotes()
    if err != nil {
        return "", err
    }
    for _, remote := range remotes {
        name := remote.Config.Name
        URLs := remote.Config().URLs
        for _, URL := range URLs {
            if URL == remoteURL {
                // Return the first remote that matches
                return name, nil
            }
        }
    }
    return "", git.ErrRemoteNotFound
}

// Generates a random remote name of length 8 
// Returns remotename, error. 
func(r *Repository) CreateRemote() (remoteName string, error) { // Refactor dependencies
    remoteName = randstr.Hex(8)
    remoteURL := s.getURL()
    r, err := r.repo.CreateRemote(&config.RemoteConfig{
        Name: remoteName,
        URLs: []string{remoteURL},
    })
    if err != nil {
        return nil, err
    }
    return
}

//Convinence function to avoid checking what kind of url we have all the time
func(s *SyncConfig) getURL() (string) {
    if s.HTTPS != nil {
        return s.HTTPS
    }
    return s.SSH
}

func splitBasicAuth(basicAuth string) (username, password string, error) {
    credentialsSlice := strings.SplitN(basicAuth, ":", 1)
    if len(credentialsSlice) != 2 {
        return "", "", ErrInvalidBasicAuth
    }
    return username, password = credentialsSlice[0], credentialsSlice[1], nil
}

func CloneHTTPS(remoteURL string, branch string, basicAuth string,) (error) {
    username, password, err := splitBasicAuth(basicAuth)
    if err != nil {
        return err
    }
    _, err := git.PlainClone(RepoPath, false, &git.CloneOptions{
        URL: remoteURL,
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

// Clones a repository using ssh url formatting and a valid sshKey read as byte slice
// Returns error if unable to clone the specified repository url
func CloneSSH(remoteURL, branch string, sshKey []byte) (error) {
    publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
    if err != nil {
        return err
    }
    _, err = git.PlainClone(RepoPath, false, &git.CloneOptions {
        URL: s.SSH,
        Progress: os.Stdout,
        RemoteName: branch,
        Auth: publicKey,
   })
   if err != nil {
        return err
   }
   return nil
}

// Validates dotsyncs configuration and sets default values if not available
// returns an descriptive error if a the config can not be validated
func(s *SyncConfig) ValidateAndSetDefaults() (error) {
    // Check that a git url has been supplied
    if s.HTTPS == nil && s.SSH == nil {
        return ErrNoRemoteURL
    }
    // Check that we don't both have ssh and https url 
    if s.HTTPS == nil && s.SSH == nil {
        return errors.New("Multiple git urls")
    }
    // Check that we have credentials if HTTPS is supplied
    if s.HTTPS != nil && s.Credentials == nil{
        return errors.New("Missing path to credentials file")
    }
    // Check that either credentials points to a valid file or that
    // the default ~/.ssh/id_rsa exists
    if s.SSH != nil {
        if s.Credentials != nil {
            if _, err := os.Stat(s.Credentials); os.IsNotExist(err) {
                return err
            }
        } else {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                return err
            }
            defaultSSHKey := filepath.Join(homeDir, ".ssh", "id_rsa")
            if _, err := os.Stat(defaultSSHKey); os.IsNotExist(err) {
                return err
            }
            s.Credentials = defaultSSHKey
        }
    }
    // Branch name defaults to main if nothing else is set
    if s.Branch == nil {
        s.Branch = "main"
    }
    // Check that we atleast have one file that we can watch
    // otherwise no point of the program continuing after a validate
    if s.Files == nil || len(s.Files) == 0 {
        return errors.New("No files to watch")
    }
}


// If remote is considered to be the single point of truth this function collects
// all files which have been changed locally and resets them to the state of the remote
// files
// Input a slice of files which are out of sync
// Returns an error if unable to sync the local files with remote
func SyncLocal(filesNotInSync []string) (error){
    err := s.UpdateOrigin()
    if err != nil {
        return err
    }
    for _, file := range filesNotInSync {
        _, fileName := filepath.Split(file)
        remoteFile := filePath.Join(RepoPath, fileName)
        err := copyFile(remoteFile, file); err != nil {
            // Log error, must likely related to file permission
        }
    }
}

// If local is considered the single point of truth this function collects all
// files which are have been changed and syncs it with remote.
// Input a slice of files which are out of sync
// Returns an error if unable to sync remote
func SyncRemote(filesNotInSync []string) (error){
    err := s.UpdateOrigin()
    if err != nil {
        return err
    }
    repository, err := git.PlainOpen(RepoPath)
    if err != nil {
        return err
    }
    workTree, err := repository.WorkTree()
    if err != nil {
        return err
    }
    filesAdded := make([]string, 0)
    for _, file : range filesNotInSync {
        _, fileName := filepath.Split(file)
        remoteFile := filePath.Join(RepoPath, fileName)
        if err := copyFile(file, remoteFile); err == nil {
            err := workTree.Add(fileName)
            if err == nil {
                filesAdded = append(filesAdded, fileName)
            } else {
                // Log error
            }
        } else {
            // Log error must likely related to file permission
        }
    }
    // Construct commit message
    commitMessage := "dotsync automated message. Updated the following files: "
    for i, fileName := range filesAdded {
        if i < len(filesAdded) -1 {
            commitMessage += fileName + ", "
        } else {
            commitMessage += fileName
        }
    }
    commit, err workTree.Commit(commitMessage, &git.CommitOptions{
        Author: &object.Signature{
            Name: "dotsync"
        },
    })
    if err != nil {
        return err
    }
    // Fix Auth
    // If this function doesnt return an error we can have a separate push function so that logic is not handled here, 
    // thus we do not need this function to be a method
    repository.Push(&git.PushOptions{
    })
}

// TODO: Refactor everything
func forSanity() {
    s := NewSynConfig()
    r := NewRepository(s)
}
