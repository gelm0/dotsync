package dotsyn

import (
    "os"
    "io"
    "io/ioutil"
    "strings"
    "path/filepath"

    diff "github.com/sergi/go-diff/diffmatchpatch"
    "gopkg.in/yaml.v2"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/transport/http"
    "github.com/thankpk/randstr" // Random string package
)

type SyncConfig struct {
	HTTPS       string      `yaml:"https,omitempty"`
	SSH         string      `yaml:"ssh,omitempty"`
	Credentials string      `yaml:"credentials,omitempty"`
	Branch      string      `yaml:"branch,omitempty"`
	Files       []string    `yaml:"files"`
}

type Repository struct {
    r *git.Repository
}

type GitOperations interface {
    Commit(commitMessage string) error
    Push() error
    CloneSSH(remoteURL string, sshKey []byte, branch string) error
    CloneHTTPS(remoteURL string, basicAuth string, branch string) error
    Pull() error
}

const (
    RepoPath = "/tmp/dotsync"
    ErrNoCredentialsFile = errors.new("Missing credentials file")
    ErrInvalidBasicAuth = errors.New("Can't extract username, password")
    ErrNoRemoteURL = errors.new("Missing remote url")
)

func copyFile(source string, destination) (error) {
    input, err := ioutil.ReadFile(source)
    if err != nil {
        return err
    }
    err := ioutil.WriteFile(destination, input, 0644)
    if err != nil {
        return err
    }
}

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
    var configPath
    if path == "" {
        homeDir, err := os.UserHomeDir()
        err != nil {
            return nil, err
        }
        // Default path is assumed to be ~/.dotsync/dotsync.yaml
        configPath = filepath.Join(homeDir, ".dotsync", "dotsync.yaml")
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

//Convinence function to avoid checking what kind of url we have all the time
func(s *SyncConfig) getURL() (string) {
    if s.HTTPS != nil {
        return s.HTTPS
    }
    return s.SSH
}

func CloneHTTPS(remoteURL string, basicAuth string, branch string) (error) {
    credentialsSlice := strings.SplitN(basicAuth, ":", 1)
    if len(credentialsSlice) != 2 {
        return ErrInvalidBasicAuth
    }
    username, password := credentialsSlice[0], credentialsSlice[1]
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

func() CloneSSH(remoteURL string, sshKey []byte) (error) {
    publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
    if err != nil {
        return err
    }
    _, err = git.PlainClone(RepoPath, false, &git.CloneOptions {
        URL: s.SSH,
        Progress: os.Stdout,
        RemoteName: s.Branch,
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

//Retrieves the remote name specified by the SyncConfigs url
// if a remote name cannot be found an error will be returned
func (r *Repository) GetRemoteName() (remoteName string, error) { //Refactor dependencies interface
    remotes, err := r.Remotes()
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
func(s *SyncConfig) CreateRemote(r *git.Repository) (remoteName string, error) { // Refactor dependencies
    remoteName = randstr.Hex(8)
    remoteURL := s.getURL()
    r, err := r.CreateRemote(&config.RemoteConfig{
        Name: remoteName,
        URLs: []string{remoteURL},
    })
    if err != nil {
        return nil, err
    }
    return
}

// Will clone repository with HTTPS url and if not available will
// assume that SyncConfig contains SSH url and use that.
// Returns error if unable to clone the specified repository url

func(s *SyncConfig) Push() (error) {

}

func(s *SyncConfig) pushHTTPS() (error) {

}

func(s *SyncConfig) pushSSH() (error) {

}
func Commit(commitMessage string) {

}


func(r *Repository) UpdateOrigin() (error) {
    // Check that directory exists and is git initialized
    // if exists -> fetch -> pull newest origin
    // Check that the repo exists and is set to correct url
    if _, err := os.Stat(repoPath + / + ".git"); os.IsNotExist(err) {
        err := s.Clone()
    // Some repository exists. Compare that we are fetching from the specified one
    } else {
        r, err := git.PlainOpen(RepoPath)
        if err != nil {
            return err
        }
        remoteName, err := s.GetRemoteName(r)
        if err != nil && err != ErrRemoteNotFound{
            return err
        }
        if remoteName == "" {
            remoteName, err = CreateRemote(r)
            if err != nil {
                return err
            }
        }
        w, err := r.WorkTree()
        if err != nil {
            return err
        }
        err := w.Pull(&git.PullOptions{RemoteName: remoteName})
        if err != nil {
            return err
        }
    }
    return nil
}

// Checks if there is a difference between local and remote files that are
// being watched
// Returns a slice of files that are not in sync with remote
func(s *SyncConfig) DiffRemoteWithLocal() (filesNotInSync []string, error) {
    err := s.UpdateOrigin()
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
