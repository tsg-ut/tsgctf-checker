package checker

import (
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Enumerate directories under challs_dir.
func EnumerateChallenges(challs_dir string) ([]string, error) {
	pathes := make([]string, 0)
	files, err := os.ReadDir(challs_dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			path, _ := filepath.Abs(filepath.Join(challs_dir, f.Name()))
			pathes = append(pathes, path)
		}
	}

	return pathes, nil
}

type asyncTestResult struct {
	executer Executer
	result   TestResult
}

// Run a test with timeout.
func run_test(executer Executer, ch chan<- asyncTestResult) {
	timeout := executer.chall.Timeout
	if timeout < 0 {
		timeout = math.MaxFloat64
	}

	res_chan := make(chan TestResult)
	killer_chan := make(chan bool)
	go executer.ExecuteDockerTest(res_chan, killer_chan)

	res := ResultRunning

	for res == ResultRunning {
		select {
		case result := <-res_chan:
			res = result
		case <-time.After(time.Duration(timeout) * time.Second):
			close(killer_chan)
			res = <-res_chan
		}
	}

	ch <- asyncTestResult{
		executer: executer,
		result:   res,
	}
}

// Run all tests using a given configuration, and record the results.
func RunRecordTests(logger *zap.SugaredLogger, conf CheckerConfig, db *sqlx.DB) error {
	if db == nil {
		logger.Error("DB is nil")
		return nil
	}

	// enumerate challenges
	chall_pathes, err := EnumerateChallenges(conf.ChallsDir)
	if err != nil {
		logger.Errorw("Failed to enumerate challenges", "error", err)
		return err
	}
	challs := make([]Challenge, 0)
	for _, path := range chall_pathes {
		chall, err := ParseChallenge(path)
		if err != nil {
			if conf.SkipNonExist {
				continue
			} else {
				logger.Errorw("Failed to parse challenge", "error", err)
				return err
			}
		}
		challs = append(challs, chall)
	}

	executers_wait_queue := make([]Executer, 0)
	num_running := 0
	result_chans := make(chan asyncTestResult, len(challs))

	// instantiate executers
	for _, chall := range challs {
		executer := Executer{
			challenge_dir: chall.SolverDir,
			chall:         chall,
			logger:        logger,
		}
		executers_wait_queue = append(executers_wait_queue, executer)
	}

	// initial runs
	for conf.ParallelNum > uint(num_running) && len(executers_wait_queue) > 0 {
		executer := executers_wait_queue[0]
		executers_wait_queue = executers_wait_queue[1:]
		go run_test(executer, result_chans)
		num_running++
	}

	// watch channel
	for result := range result_chans {
		num_running--

		for conf.ParallelNum > uint(num_running) && len(executers_wait_queue) > 0 {
			executer := executers_wait_queue[0]
			executers_wait_queue = executers_wait_queue[1:]
			go run_test(executer, result_chans)
			num_running++
		}

		if err := RecordResult(db, result.executer.chall, result.result); err != nil {
			logger.Errorw("Failed to record result", "error", err)
			close(result_chans)
			return err
		}

		if num_running == 0 && len(executers_wait_queue) == 0 {
			close(result_chans)
		}
	}

	return nil
}
