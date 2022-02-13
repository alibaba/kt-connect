package util

import (
	"crypto/rsa"
	"github.com/stretchr/testify/require"
	"math/big"
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
			_ = os.Remove(tt.args.privateKeyPath)
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
	}{
		{
			name: "should generate rsa key",
			args: args{bitSize: 2048},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generatePrivateKey(tt.args.bitSize)
			require.Empty(t, err, "generatePrivateKey() error = %v", err)
			require.Equal(t, 256, got.PublicKey.Size(), "generatePrivateKey() = %v, want %v", got.PublicKey.Size(), 256)
		})
	}
}

func Test_encodePublicKey(t *testing.T) {
	type args struct {
		publicKey *rsa.PublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should encode public key",
			args: args{publicKey: &rsa.PublicKey{
				N: big.NewInt(7456871076443426339),
				E: 65537},
			},
			want: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAACGd8ILxfS44j\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodePublicKey(tt.args.publicKey)
			require.Equal(t, err != nil, tt.wantErr, "encodePublicKey() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, string(got), tt.want, "encodePublicKey() = %v, want %v", got, tt.want)
		})
	}
}
