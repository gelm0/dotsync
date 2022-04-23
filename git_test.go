package dotsync

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

var workingConfig = SyncConfig{
	URL:     "http://test.com",
	KeyFile: ".ssh/id_rsa",
	Branch:  "main",
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

func createMockRepository() afero.Fs {
	appFs := afero.NewMemMapFs()
	appFs.MkdirAll(filepath.Join(RepoPath, ".git"), 0755)
	appFs.MkdirAll(".ssh", 0755)
	privateKey, publicKey := generateSSHKeys()
	afero.WriteFile(appFs, ".ssh/id_rsa", privateKey, 0600)
	afero.WriteFile(appFs, ".ssh/id_rsa.pub", publicKey, 0600)
	return appFs
}

type mockGitExtension struct{}

func (m *mockGitExtension) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	fs := afero.NewOsFs()
	name, err := afero.TempDir(fs, "", "")
	if err != nil {
		panic(err)
	}
	return git.PlainInit(name, true)
}

func (m *mockGitExtension) plainOpen(path string) (*git.Repository, error) {
	fs := afero.NewOsFs()
	name, err := afero.TempDir(fs, "", "")
	if err != nil {
		panic(err)
	}
	return git.PlainInit(name, true)

}

func TestNewRepositoryOpens(t *testing.T) {
	m := &mockGitExtension{}
	appFs := createMockRepository()
	_, err := NewRepository(workingConfig, appFs, m)
	assert.NoError(t, err)
}

func TestNewRepositoryThrowsErrorWithBadKey(t *testing.T) {

}

func TestNewRepositoryClonesWhenEmpty(t *testing.T) {

}
