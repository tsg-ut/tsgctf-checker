package checker

// This file implements the actual execution of health checks.

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
)

// Executer of a single test.
type Executer struct {
	challenge_dir string
	logger        *zap.SugaredLogger
	chall         Challenge
}

// Test result
type TestResult int

const (
	// Test passed
	ResultSuccess TestResult = iota
	// Test timed out
	ResultTimeout
	// Test failed to execute
	ResultExecutionFailure
	// Test interrupted by signal
	ResultTestInterrupted
	// Test failed
	ResultFailure
	// Test running
	ResultRunning
)

func (tr TestResult) ToMessage() string {
	switch tr {
	case ResultSuccess:
		return "Solvable"
	case ResultTimeout:
		return "Timeout"
	case ResultExecutionFailure, ResultTestInterrupted, ResultFailure:
		return "Unsolvable"
	case ResultRunning:
		return "Running"
	default:
		return "Unknown"
	}
}

func (tr TestResult) ToColor() string {
	switch tr {
	case ResultSuccess:
		return "33FF99"
	case ResultTimeout:
		return "900000"
	case ResultExecutionFailure, ResultTestInterrupted, ResultFailure:
		return "CC0000"
	case ResultRunning:
		return "C0C0C0"
	default:
		return "C0C0C0"
	}
}

func (e *Executer) check_before_execution() error {
	// check if Dockerfile exists
	if _, err := os.Stat(filepath.Join(e.chall.SolverDir, "Dockerfile")); os.IsNotExist(err) {
		return fmt.Errorf("[%s] Dockerfile not found in %s", e.chall.Name, e.challenge_dir)
	}

	return nil
}

// Execute a test using a Dockerfile.
// This function is blocked until the subprocess finishes.
// The caller can get the test result from res_chan.
// The caller can kill the subprocess by sending a signal to killer_chan and it returns ResultTimeout.
// This function is notified if SIGTERM or SIGINT is sent to the process,
// and it cleans up subprocess and returns ResultTestInterrupted.
// Note that it sends ResultRunning to res_chan when it starts executing the test.
func (e *Executer) ExecuteDockerTest(res_chan chan TestResult, killer_chan <-chan bool) {
	if err := e.check_before_execution(); err != nil {
		e.logger.Errorf("[%s] Failed to execute test: \n%w", e.chall.Name, err)
		res_chan <- ResultExecutionFailure
		return
	}

	// prepare command
	chall := e.chall
	container_name := fmt.Sprintf("container_solver_%s", chall.Name)
	image_name := fmt.Sprintf("solver_%s", chall.Name)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run --name %s --rm $(docker build -qt solver_%s %s)", container_name, image_name, chall.SolverDir))

	// termination signal hook
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// execute test async
	res_chan_internal := make(chan error)
	if err := cmd.Start(); err != nil {
		e.logger.Warnf("[%s] Failed to start test: \n%w", chall.Name, err)
		res_chan <- ResultFailure
		return
	}
	res_chan <- ResultRunning
	e.logger.Infof("[%s] Test started as pid %d in %s.", chall.Name, cmd.Process.Pid, container_name)
	go func() {
		res_chan_internal <- cmd.Wait()
	}()

	cleanup_container := func() {
		// remove container
		if err := exec.Command("docker", "rm", "-f", container_name).Run(); err != nil {
			e.logger.Errorf("[%s] Failed to remove container(%s):\n%v", chall.Name, container_name, err)
		}
		// check if process is running
		// kill process
		if err := cmd.Process.Kill(); err != nil {
			e.logger.Errorf("[%s] Failed to kill process: %v", chall.Name, err)
		}
	}

	// wait for result
	select {
	// checker process terminated by signal
	case <-signal_chan:
		e.logger.Infof("[%s] Checker process interrupted, cleaning up docker container...", chall.Name)
		cleanup_container()
		res_chan <- ResultTestInterrupted
		break
	// timeout
	case _, ok := <-killer_chan:
		if !ok {
			cleanup_container()
		}
		res_chan <- ResultTimeout
		break
	// test finished
	case err := <-res_chan_internal:
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				e.logger.Infof("[%s] Test failed with status %d", chall.Name, exiterr.ExitCode())
				res_chan <- ResultFailure
			}
		} else {
			// test ends without any failure
			e.logger.Infof("[%s] exits with status code 0.", chall.Name)
			res_chan <- ResultSuccess
		}
		break
	}
}
