package checker

// This file implements the actual execution of health checks.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

// Executer of a single test.
type Executer struct {
	challenge_dir string
	logger        *zap.SugaredLogger
	chall         Challenge
}

type TestResultMessage struct {
	Result TestResult
	Stdout string
	Errlog string
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
func (e *Executer) ExecuteDockerTest(res_chan chan TestResultMessage, killer_chan <-chan bool, conf CheckerConfig) {
	if err := e.check_before_execution(); err != nil {
		e.logger.Errorf("[%s] Failed to execute test: \n%w", e.chall.Name, err)
		res_chan <- TestResultMessage{Result: ResultExecutionFailure, Errlog: err.Error()}
		return
	}

	// prepare command
	chall := e.chall
	container_name := fmt.Sprintf("container_solver_%s", strings.ToLower(chall.Name))
	image_name := fmt.Sprintf("solver_%s", strings.ToLower(chall.Name))
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run %s --name %s --rm $(docker build -qt %s %s) %s %d", conf.ExtraDockerArg, container_name, image_name, chall.SolverDir, chall.target.Host, chall.target.Port))

	var errbuf bytes.Buffer
	var outbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	// termination signal hook
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// execute test async
	res_chan_internal := make(chan error)
	if err := cmd.Start(); err != nil {
		e.logger.Warnf("[%s] Failed to start test: \n%w", chall.Name, err)
		res_chan <- TestResultMessage{ResultFailure, outbuf.String(), err.Error()}
		return
	}
	res_chan <- TestResultMessage{ResultRunning, outbuf.String(), errbuf.String()}
	e.logger.Infof("[%s] Test started as pid %d in %s.", chall.Name, cmd.Process.Pid, container_name)
	go func() {
		res_chan_internal <- cmd.Wait()
	}()

	cleanup_container := func() {
		// check if process is running
		// kill process
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			if err.Error() != "os: process already finished" {
				e.logger.Errorf("[%s] Failed to kill process: %v", chall.Name, err)
			}
		}
		// remove container
		if err := exec.Command("docker", "stop", container_name).Run(); err != nil {
			e.logger.Errorf("[%s] Failed to remove container (%s):\n%v", chall.Name, container_name, err)
		}
	}

	// wait for result
	select {
	// checker process terminated by signal
	case <-signal_chan:
		e.logger.Infof("[%s] Checker process interrupted, cleaning up docker container...", chall.Name)
		cleanup_container()
		res_chan <- TestResultMessage{ResultTestInterrupted, "", "Interrupted by signal."}
		break
	// timeout
	case <-killer_chan:
		e.logger.Infof("[%s] Test timed out. Stopping container.", chall.Name)
		cleanup_container()
		e.logger.Infof("[%s] Container stopped.", chall.Name)
		if conf.Vervose {
			e.logger.Infof("[%s] stdout: %s", chall.Name, outbuf.String())
			e.logger.Infof("[%s] stderr: %s", chall.Name, errbuf.String())
		}
		res_chan <- TestResultMessage{ResultTimeout, outbuf.String(), errbuf.String()}
		break
	// test finished
	case err := <-res_chan_internal:
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				e.logger.Infof("[%s] Test failed with status %d", chall.Name, exiterr.ExitCode())
				if conf.Vervose {
					e.logger.Infof("[%s] stdout: %s", chall.Name, outbuf.String())
					e.logger.Infof("[%s] stderr: %s", chall.Name, errbuf.String())
				}
				res_chan <- TestResultMessage{ResultFailure, outbuf.String(), errbuf.String()}
			}
		} else {
			// test ends without any failure
			e.logger.Infof("[%s] exits with status code 0.", chall.Name)
			res_chan <- TestResultMessage{ResultSuccess, "", ""}
		}
		break
	}
}
