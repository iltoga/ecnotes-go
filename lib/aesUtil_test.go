package lib

import (
	"testing"
)

// func main() {
// 	CIPHER_KEY := []byte("0123456789012345")
// 	msg := "A quick brown fox jumped over the lazy dog."

// 	if encrypted, err := Encrypt(CIPHER_KEY, msg); err != nil {
// 		log.Println(err)
// 	} else {
// 		log.Printf("CIPHER KEY: %s\n", string(CIPHER_KEY))
// 		log.Printf("ENCRYPTED: %s\n", encrypted)

// 		if decrypted, err := Decrypt(CIPHER_KEY, encrypted); err != nil {
// 			log.Println(err)
// 		} else {
// 			log.Printf("DECRYPTED: %s\n", decrypted)
// 		}
// 	}
// }

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
			_, err := Encrypt(tt.args.key, tt.args.message)
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
				securemess: "emqCN1rO7DhJSw0L_0NJJYxiDwgIC7s_LYTwDXKlDIxwvnHvLGOi12ft0qwulgPwue-glzoPikL_m7Y=",
			},
			wantDecodedmess: "A quick brown fox jumped over the lazy dog.",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecodedmess, err := Decrypt(tt.args.key, tt.args.securemess)
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
