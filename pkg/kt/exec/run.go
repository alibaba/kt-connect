package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd, name string, debug bool) (err error) {
	runCmd(cmd, name, debug)
	err = cmd.Wait()
	return
}

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string, debug bool) (err error) {
	err = runCmd(cmd, name, debug)
	if err != nil {
		return
	}
	go func() {
		err = cmd.Wait()
		log.Info().Msgf("%s finished", name)
	}()
	return
}

func runCmd(cmd *exec.Cmd, name string, debug bool) (err error) {
	log.Debug().Msgf("Child, os.Args = %+v", os.Args)
	log.Debug().Msgf("Child, name = %s, cmd.Args = %+v", name, cmd.Args)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	// var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	go func() {
		if stdoutIn != nil {
			_, _ = io.Copy(stdout, stdoutIn)
		}
	}()
	go func() {
		if stderrIn != nil {
			_, _ = io.Copy(stderr, stderrIn)
		}
	}()

	err = cmd.Start()

	if err != nil {
		return
	}

	time.Sleep(time.Duration(1) * time.Second)
	pid := cmd.Process.Pid
	log.Info().Msgf("%s start at pid: %d", name, pid)
	return
}
