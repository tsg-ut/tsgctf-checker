package badge

import (
	"testing"
	"time"

	"github.com/tsg-ut/tsgctf-checker/checker"
)

type Challenge = checker.Challenge

func TestBadge_GetBadge(t *testing.T) {
	type args struct {
		chall_name string
		result     checker.TestResult
		want       string
	}
	tests := []args{
		{
			chall_name: "just-success",
			result:     checker.ResultSuccess,
			want:       "https://img.shields.io/badge/Solvable-10/15_13:28:33_UTC-33FF99",
		},
		{
			chall_name: "just-fail",
			result:     checker.ResultFailure,
			want:       "https://img.shields.io/badge/Unsolvable-10/15_13:28:33_UTC-CC0000",
		},
	}

	const_time := time.Date(2023, 10, 15, 13, 28, 33, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.chall_name, func(t *testing.T) {
			badge, err := GetBadge(tt.chall_name, tt.result, const_time)
			if err != nil {
				t.Errorf("GetBadge() error = %v", err)
			}
			if badge != tt.want {
				t.Errorf("GetBadge() got = %v, want %v", badge, tt.want)
			}
		})
	}
}
