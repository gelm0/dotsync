package dotsync

import (
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"encoding/pem"
	"testing"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var workingConfig = SyncConfig{
	URL: "http://test.com",
	KeyFile: ".ssh/id_rsa",
	Branch: "main",
	Files: []string{},
}

func generateSSHKeys() ([]byte, []byte) {
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

func TestNewRepositoryOpens(t *testing.T) {
	appFs := createMockRepository()
	_, err := NewRepository(workingConfig, appFs)
	assert.NoError(t, err)
}

func TestNewRepositoryThrowsErrorWithBadKey(t *testing.T) {

}

func TestNewRepositoryClonesWhenEmpty(t *testing.T) {

}