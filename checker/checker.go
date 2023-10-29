package checker

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func enumerateGenreChallenges(genre_dir string) ([]string, error) {
	pathes := make([]string, 0)
	files, err := os.ReadDir(genre_dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			_, err := os.Stat(filepath.Join(genre_dir, f.Name(), "solver/info.json"))
			if err != nil {
				continue
			}
			path, _ := filepath.Abs(filepath.Join(genre_dir, f.Name()))
			pathes = append(pathes, path)
		}
	}

	return pathes, nil
}

// Enumerate directories under challs_dir.
func EnumerateChallenges(challs_dir string, have_genre_dir bool) ([]string, error) {
	if have_genre_dir {
		pathes := make([]string, 0)
		genre_dirs, err := os.ReadDir(challs_dir)
		if err != nil {
			return nil, err
		}
		for _, genre_dir := range genre_dirs {
			if genre_dir.IsDir() {
				genre_challs, err := enumerateGenreChallenges(filepath.Join(challs_dir, genre_dir.Name()))
				if err != nil {
					return nil, err
				}
				pathes = append(pathes, genre_challs...)
			}
		}
		return pathes, nil
	} else {
		return enumerateGenreChallenges(challs_dir)
	}
}

type asyncTestResult struct {
	executer Executer
	result   TestResultMessage
}

// Run a test with timeout.
func run_test(executer Executer, ch chan<- asyncTestResult, conf CheckerConfig) {
	timeout := executer.chall.Timeout
	if timeout < 0 {
		timeout = math.MaxFloat64
	}

	res_chan := make(chan TestResultMessage)
	killer_chan := make(chan bool)
	go executer.ExecuteDockerTest(res_chan, killer_chan, conf)

	res := TestResultMessage{ResultRunning, "", ""}

	for res.Result == ResultRunning {
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

func parseTargets(logger *zap.SugaredLogger, path string) ([]Target, error) {
	targets := make([]Target, 0)
	targets_file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer targets_file.Close()

	reader := csv.NewReader(targets_file)
	reader.Comma = ','
	reader.Comment = '#'
	reader.FieldsPerRecord = 3
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		port, err := strconv.Atoi(row[2])
		if err != nil {
			return nil, err
		}
		targets = append(targets, Target{
			ChallengeName: row[0],
			Host:          row[1],
			Port:          port,
		})
	}

	return targets, nil
}

// Run all tests using a given configuration, and record the results.
func RunRecordTests(logger *zap.SugaredLogger, conf CheckerConfig, db *sqlx.DB) error {
	slack_notifier := NewSlackNotifier(conf.SlackToken, conf.SlackChannel, logger)

	if conf.Dryrun == false && db == nil {
		logger.Error("DB is nil")
		return nil
	}

	// read targets
	targets, err := parseTargets(logger, conf.TargetsFile)
	if err != nil {
		logger.Errorw(fmt.Sprintf("Failed to parse targets: %s", conf.TargetsFile), "error", err)
		return err
	}

	// enumerate challenges
	chall_pathes, err := EnumerateChallenges(conf.ChallsDir, conf.HaveGenreDir)
	if err != nil {
		logger.Errorw("Failed to enumerate challenges", "error", err)
		return err
	}
	logger.Infof("Found %d challenges", len(chall_pathes))
	if len(chall_pathes) == 0 {
		logger.Info("No challenges found")
		return nil
	}

	challs := make([]Challenge, 0)
	for _, path := range chall_pathes {
		chall, err := ParseChallenge(path, targets)
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
		go run_test(executer, result_chans, conf)
		num_running++
	}

	// watch channel
	for result := range result_chans {
		num_running--

		for conf.ParallelNum > uint(num_running) && len(executers_wait_queue) > 0 {
			executer := executers_wait_queue[0]
			executers_wait_queue = executers_wait_queue[1:]
			go run_test(executer, result_chans, conf)
			num_running++
		}

		if conf.Dryrun == false {
			if err := RecordResult(db, result.executer.chall, result.result.Result); err != nil {
				logger.Errorw("Failed to record result", "error", err)
				close(result_chans)
				return err
			}

			if conf.NotifySlack && result.result.Result != ResultSuccess {
				slack_notifier.NotifyError(result.executer.chall, result.result.Result, result.result.Stdout, result.result.Errlog)
			}
		}

		if num_running == 0 && len(executers_wait_queue) == 0 {
			close(result_chans)
		}
	}

	return nil
}
