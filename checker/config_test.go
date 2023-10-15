package checker

import (
	"os"
	"testing"
)

func TestConfig_ReadConf(t *testing.T) {
	cwd := testing_cd_root(t)
	defer os.Chdir(cwd)

	type args struct {
		config_path string
	}
	tests := []struct {
		name    string
		args    args
		want    CheckerConfig
		wantErr bool
	}{
		{
			name: "normal-config",
			args: args{
				config_path: "tests/assets/config.json",
			},
			want: CheckerConfig{
				ParallelNum: 10,
				ChallsDir:   "tests/assets/challs",
				Retries:     0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := ReadConf(tt.args.config_path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.ParallelNum != tt.want.ParallelNum {
				t.Errorf("ReadConf() got = %v, want %v", got, tt.want)
			}
			if got.ChallsDir != tt.want.ChallsDir {
				t.Errorf("ReadConf() got = %v, want %v", got, tt.want)
			}
			if got.Retries != tt.want.Retries {
				t.Errorf("ReadConf() got = %v, want %v", got, tt.want)
			}
		})
	}
}
