package dotsync

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"testing"
)

var workingConfig = SyncConfig{
	GitConfig: GitConfig {
		URL:     "http://test.com",
		KeyFile: ".ssh/id_rsa",
		Branch:  "main",
	},
	Files:   []string{},
}

func generateSSHKeys() ([]byte, []byte) {
	// Keep the bytes low to speed up test
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic("Not able to generate private key")
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic("Not able to generate public key")
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privateBytes := pem.EncodeToMemory(&privBlock)
	publicBytes := ssh.MarshalAuthorizedKey(publicKey)
	return privateBytes, publicBytes
}

func createTempSSSHDir() afero.Fs {
	appFs := afero.NewMemMapFs()
	appFs.MkdirAll(".ssh", 0755)
	privateKey, publicKey := generateSSHKeys()
	afero.WriteFile(appFs, ".ssh/id_rsa", privateKey, 0600)
	afero.WriteFile(appFs, ".ssh/id_rsa.pub", publicKey, 0600)
	return appFs
}

type mockGitExtension struct {
	plainCloneCalled int
	plainOpenCalled  int
}

func (m *mockGitExtension) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	m.plainCloneCalled += 1
	return &git.Repository{}, nil
}

func (m *mockGitExtension) plainOpen(path string) (*git.Repository, error) {
	m.plainOpenCalled += 1
	return &git.Repository{}, nil
}

func TestNewRepositoryOpens(t *testing.T) {
	m := &mockGitExtension{
		plainCloneCalled: 0,
		plainOpenCalled:  0,
	}
	sshFs := createTempSSSHDir()
	sshFs.MkdirAll(filepath.Join(RepoPath, ".git"), 0755)
	r, err := newRepository(workingConfig, sshFs, m)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 1, m.plainOpenCalled)
}

func TestNewRepositoryThrowsErrorWithBadKey(t *testing.T) {
	m := &mockGitExtension{
		plainCloneCalled: 0,
		plainOpenCalled:  0,
	}
	sshFs := createTempSSSHDir()
	afero.WriteFile(sshFs, ".ssh/id_rsa", []byte{}, 0600)
	_, err := newRepository(workingConfig, sshFs, m)
	assert.Error(t, err)
}

func TestNewRepositoryClonesWhenEmpty(t *testing.T) {
	m := &mockGitExtension{
		plainCloneCalled: 0,
		plainOpenCalled:  0,
	}
	sshFs := createTempSSSHDir()
	r, err := newRepository(workingConfig, sshFs, m)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 1, m.plainCloneCalled)
}
