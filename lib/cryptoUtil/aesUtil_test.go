package cryptoUtil_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	type args struct {
		key     []byte
		message []byte
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
				message: []byte("A quick brown fox jumped over the lazy dog."),
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
	msg, _ := hex.DecodeString("73a95593798a6717658eda94d9db94c7f90a58ba750a214ee9be22e9eb724da0081213843543f0f2a5359b236e1591c46928e92ad314d29723ec6970cd3a24cef44df87caad907")
	type args struct {
		key        []byte
		securemess []byte
	}
	tests := []struct {
		name            string
		args            args
		wantDecodedmess []byte
		wantErr         bool
	}{
		{
			name: "TestEncrypt_01",
			args: args{
				key:        []byte("0123456789012345"),
				securemess: msg,
			},
			wantDecodedmess: []byte{65, 32, 113, 117, 105, 99, 107, 32, 98, 114, 111, 119, 110, 32, 102, 111, 120, 32, 106, 117, 109, 112, 101, 100, 32, 111, 118, 101, 114, 32, 116, 104, 101, 32, 108, 97, 122, 121, 32, 100, 111, 103, 46},
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
			assert.Equal(t, tt.wantDecodedmess, gotDecodedmess, fmt.Sprintf("Decrypt() = %v, want %v", gotDecodedmess, tt.wantDecodedmess))
		})
	}
}
