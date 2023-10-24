package checker

import (
	"os"
	"testing"
)

func TestChallenge_ParseChallenge(t *testing.T) {
	cwd := testing_cd_root(t)
	defer os.Chdir(cwd)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    Challenge
		wantErr bool
	}{
		{
			name: "just-success",
			args: args{
				path: "tests/assets/challs/just-success",
			},
			want: Challenge{
				Name:    "just-success",
				Timeout: 60,
			},
			wantErr: false,
		},
		{
			name: "just-fail",
			args: args{
				path: "tests/assets/challs/just-fail",
			},
			want: Challenge{
				Name:    "just-fail",
				Timeout: 60,
			},
			wantErr: false,
		},
		{
			name: "just-success-long",
			args: args{
				path: "tests/assets/challs/just-success-long",
			},
			want: Challenge{
				Name:    "just-success-long",
				Timeout: 5,
			},
			wantErr: false,
		},
	}

	targets := []Target{
		{
			ChallengeName: "just-success",
			Host:          "localhost",
			Port:          3306,
		},
		{
			ChallengeName: "just-fail",
			Host:          "localhost",
			Port:          3306,
		},
		{
			ChallengeName: "just-success-long",
			Host:          "localhost",
			Port:          3306,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChallenge(tt.args.path, targets)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChallenge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("ParseChallenge() got = %v, want %v", got, tt.want)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("ParseChallenge() got = %v, want %v", got, tt.want)
			}
		})
	}
}
