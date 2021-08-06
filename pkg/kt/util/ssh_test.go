package util

import (
	"crypto/rsa"
	"os"
	"reflect"
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.PrivateKey) == 0 {
				t.Errorf("fail generate private key")
			}
			if len(got.PublicKey) == 0 {
				t.Errorf("fail generate public key")
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("generatePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generatePrivateKey() = %v, want %v", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("generatePublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generatePublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
