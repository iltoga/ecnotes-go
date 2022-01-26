package lib

import "testing"

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
			_, err := SecureRandomStr(tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
