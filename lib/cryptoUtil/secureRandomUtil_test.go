package cryptoUtil_test

import (
	"testing"

	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
)

func TestSecureRandomStr(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				length: 10,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cryptoUtil.SecureRandomStr(tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_hash(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				s: "test string",
			},
			want: "10e6d647af44624442f388c2c14a787ff8b17e6165b83d767ec047768d8cbcb71a1a3226e7cc7816bc79c0427d94a9da688c41a3992c7bf5e4d7cc3e0be5dbac",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cryptoUtil.Hash(tt.args.s); got != tt.want {
				t.Errorf("hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
