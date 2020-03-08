package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

const (
	bitSize = 2048
)

var (
	generator *SSHGenerator
	once      sync.Once
)

// SSHGenerator ssh-key operator
type SSHGenerator struct {
	PrivateKey, PublicKey []byte
}

// InitSSH prepare ssh keypair
func InitSSH() error {
	var err error
	once.Do(func() {
		generator = &SSHGenerator{}
		if err = generator.generate(); err != nil {
			return
		}
		err = PrepareSSHPrivateKey(generator)
	})
	return err
}

// GetSSHGenerator get ssh generator
func GetSSHGenerator() (*SSHGenerator, error) {
	if generator == nil {
		return generator, errors.New("ssh generator is nil")
	}
	return generator, nil
}

// generate generate private key & public key
func (ssh *SSHGenerator) generate() error {
	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		return err
	}
	ssh.PrivateKey = encodePrivateKeyToPEM(privateKey)
	ssh.PublicKey, err = generatePublicKey(&privateKey.PublicKey)
	return err
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		log.Error().Err(err).Msg("generate private key failed")
		return nil, err
	}
	err = privateKey.Validate()
	if err != nil {
		log.Error().Err(err).Msg("private key invalid")
		return nil, err
	}
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.SSHCM and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		log.Error().Err(err).Msg("generate public key")
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}
