package util

import (
	"crypto/rsa"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	type args struct {
		privateKeyPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "shouldGenerateSSHKeyPair",
			args: args{
				privateKeyPath: "/tmp/sshkeypair",
			},
			wantErr: false,
		},
		{
			name: "supportZNDirSSHKeyPair",
			args: args{
				privateKeyPath: "/tmp/目录/sshkeypair",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(tt.args.privateKeyPath)
			got, err := Generate(tt.args.privateKeyPath)
			require.Equal(t, err != nil, tt.wantErr, "Generate() error = %v, wantErr %v", err, tt.wantErr)
			require.NotEmpty(t, got.PrivateKey, "fail generate private key")
			require.NotEmpty(t, got.PublicKey,"fail generate public key")
		})
	}
}

func Test_generatePrivateKey(t *testing.T) {
	type args struct {
		bitSize int
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generatePrivateKey(tt.args.bitSize)
			require.Equal(t, err != nil, tt.wantErr,
				"generatePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, got, tt.want,
				"generatePrivateKey() = %v, want %v", got, tt.want)
		})
	}
}

func Test_generatePublicKey(t *testing.T) {
	type args struct {
		privatekey *rsa.PublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generatePublicKey(tt.args.privatekey)
			require.Equal(t, err != nil, tt.wantErr,
				"generatePublicKey() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, got, tt.want,
				"generatePublicKey() = %v, want %v", got, tt.want)
		})
	}
}
