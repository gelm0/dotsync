package dotsync

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

const (
    RepoPath = "/tmp/dotsync"
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
func New(path string) (*SyncConfig, error) {
    var configPath
    if path == "" {
        homeDir, err := os.UserHomeDir()
        err != nil {
            return nil, err
        }
        // Default path is assumed to be ~/.dotsync/dotsync.yaml
        configPath = homeDir + "/.dotsync/dotsync.yaml")
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

//Convinence function to avoid checking what kind of url we have all the time
func(s *SyncConfig) getURL (URL string) {
    if s.HTTPS != nil {
        URL = s.HTTPS
    } else {
        URL s.SSH
    }
    return
}

// Validates dotsyncs configuration and sets default values if not available
// returns an descriptive error if a the config can not be validated
func(s *SyncConfig) ValidateAndSetDefaults() (error) {
    // Check that a git url has been supplied
    if s.HTTPS == nil && s.SSH == nil {
        return errors.New("Missing git repository to sync")
    }
    // Check that we don't both have ssh and https url 
    if s.HTTPS == nil && s.SSH == nil {
        return errors.New("Multiple git urls")
    }
    // Check that we have credentials if HTTPS is supplied
    if s.HTTPS != nil && s.Credentials == nil{
        return errors.New("Missing usename:password in credentials")
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
            defaultSSHKey := homeDir + ".ssh/id_rsa"
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
func(s *SyncConfig) GetRemoteName(r *git.Repository) (remoteName string, error) {
    remoteURL := s.getURL()
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
func(s *SyncConfig) CreateRemote(r *git.Repository) (remoteName string, error) {
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


func(s *SyncConfig) extractBasicAuthCredentials() (string, string, error){
    // Assumes that username does not contain ":" but does allow for ":" (colon) in password
    credentials := strings.SplitN(s.Credentials, ":", 1)
    if len(credentials != 2 {
        return nil, nil, errors.New("Credentials not formatted correctly")
    }
    return credentials[0], credentials[1], nil

}

// Will clone repository with HTTPS url and if not available will
// assume that SyncConfig contains SSH url and use that.
// Returns error if unable to clone the specified repository url
func(s *SyncConfig) Clone() (error) {
    if s.HTTPS != nil  && err := s.cloneHTTPS(); err != nil {
            return err
    } else if err := s.cloneSSH(); err != nil {
        return err
    }
    return nil
}

func(s *SyncConfig) cloneHTTPS() (error) {
    username, password, err := s.extractBasicAuthCredentials()
    if err != nil {
        return err
    }
    _, err := git.PlainClone(RepoPath, false, &git.CloneOptions{
        URL: s.HTTPS,
        RemoteName: s.Branch,
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

func(s *SyncConfig) cloneSSH() (error) {
    sshKey, err := os.ReadFile(s.Credentials)
    if err != nil {
        return errors.New("Error reading sshkey, %s", err)
    }
    publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
    if err != nil {
        return errors.New("Error creating public key")
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

func(s *SyncConfig) UpdateOrigin() (error) {
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
    repository.Push(&git.PushOptions{


    })
}
