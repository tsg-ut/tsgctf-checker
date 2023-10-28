package checker

import (
	"os"
	"syscall"
	"testing"
)

func TestExecuter_ExecuteDockerTest(t *testing.T) {
	cwd := testing_cd_root(t)
	defer os.Chdir(cwd)

	type fields struct {
		challenge_dir string
	}
	type args struct {
		res_chan        chan TestResultMessage
		killer_chan     chan bool
		expected_result TestResult
		// Send kill signal using channel
		should_kill bool
		// Send SIGTERM to myself
		should_interrupt_myself bool
	}

	logger := create_logger()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "just-success",
			fields: fields{
				challenge_dir: "tests/assets/challs/just-success",
			},
			args: args{
				res_chan:                make(chan TestResultMessage),
				killer_chan:             make(chan bool),
				expected_result:         ResultSuccess,
				should_kill:             false,
				should_interrupt_myself: false,
			},
		},
		{
			name: "just-fail",
			fields: fields{
				challenge_dir: "tests/assets/challs/just-fail",
			},
			args: args{
				res_chan:                make(chan TestResultMessage),
				killer_chan:             make(chan bool),
				expected_result:         ResultFailure,
				should_kill:             false,
				should_interrupt_myself: false,
			},
		},
		{
			name: "just-long",
			fields: fields{
				challenge_dir: "tests/assets/challs/just-success-long",
			},
			args: args{
				res_chan:                make(chan TestResultMessage),
				killer_chan:             make(chan bool),
				expected_result:         ResultTimeout,
				should_kill:             true,
				should_interrupt_myself: false,
			},
		},
		{
			name: "interrupt",
			fields: fields{
				challenge_dir: "tests/assets/challs/just-success-long",
			},
			args: args{
				res_chan:                make(chan TestResultMessage),
				killer_chan:             make(chan bool),
				expected_result:         ResultTestInterrupted,
				should_kill:             false,
				should_interrupt_myself: true,
			},
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
			chall, err := ParseChallenge(tt.fields.challenge_dir, targets)
			if err != nil {
				t.Errorf("Failed to parse challenge: %v", err)
			}
			e := &Executer{
				challenge_dir: tt.fields.challenge_dir,
				chall:         chall,
				logger:        logger,
			}
			go e.ExecuteDockerTest(tt.args.res_chan, tt.args.killer_chan, CheckerConfig{})

			var res TestResultMessage
			if tt.args.should_kill {
				close(tt.args.killer_chan)
			}
			if tt.args.should_interrupt_myself {
				res = <-tt.args.res_chan // wait RESULT_RUNNING
				syscall.Kill(os.Getpid(), syscall.SIGTERM)
			}

			for {
				res = <-tt.args.res_chan
				if res.Result != ResultRunning {
					break
				}
			}
			if res.Result != tt.args.expected_result {
				t.Errorf("Expected result %d, got %d", tt.args.expected_result, res.Result)
			}
		})
	}
}
