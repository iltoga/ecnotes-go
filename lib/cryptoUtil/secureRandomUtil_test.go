package cryptoUtil_test

import (
	"bytes"
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
			wantErr: false,
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
		want []byte
	}{
		{
			name: "test1",
			args: args{
				s: "test string",
			},
			want: []byte{
				119,
				233,
				243,
				83,
				67,
				24,
				51,
				195,
				22,
				189,
				65,
				220,
				136,
				103,
				13,
				154,
				210,
				29,
				46,
				89,
				80,
				214,
				245,
				226,
				52,
				111,
				46,
				136,
				89,
				244,
				252,
				155,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cryptoUtil.Hash(tt.args.s); !bytes.Equal(got, tt.want) {
				t.Errorf("hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
