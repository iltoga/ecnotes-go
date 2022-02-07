package cryptoUtil_test

import (
	"testing"

	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
)

func TestEncrypt(t *testing.T) {
	type args struct {
		key     []byte
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestEncrypt_01",
			args: args{
				key:     []byte("0123456789012345"),
				message: "A quick brown fox jumped over the lazy dog.",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cryptoUtil.EncryptAES256(tt.args.key, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	type args struct {
		key        []byte
		securemess string
	}
	tests := []struct {
		name            string
		args            args
		wantDecodedmess string
		wantErr         bool
	}{
		{
			name: "TestEncrypt_01",
			args: args{
				key:        []byte("0123456789012345"),
				securemess: "73a95593798a6717658eda94d9db94c7f90a58ba750a214ee9be22e9eb724da0081213843543f0f2a5359b236e1591c46928e92ad314d29723ec6970cd3a24cef44df87caad907",
			},
			wantDecodedmess: "A quick brown fox jumped over the lazy dog.",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecodedmess, err := cryptoUtil.DecryptAES256(tt.args.key, tt.args.securemess)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDecodedmess != tt.wantDecodedmess {
				t.Errorf("Decrypt() = %v, want %v", gotDecodedmess, tt.wantDecodedmess)
			}
		})
	}
}
