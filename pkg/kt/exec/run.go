package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// Run run cmd
func RunAndWait(cmd *exec.Cmd, name string, debug bool) (err error) {
	runCmd(cmd, name, debug)
	err = cmd.Wait()
	return
}

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string, debug bool) (err error) {
	runCmd(cmd, name, debug)
	go func() {
		err = cmd.Wait()
		log.Printf("%s finished", name)
	}()
	return
}

func runCmd(cmd *exec.Cmd, name string, debug bool) (err error) {
	log.Debug().Msgf("Child, os.Args = %+v", os.Args)
	log.Debug().Msgf("Child, cmd.Args = %+v", cmd.Args)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Start()

	if err != nil {
		return
	}

	time.Sleep(time.Duration(2) * time.Second)
	pid := cmd.Process.Pid
	log.Printf("%s start at pid: %d", name, pid)
	return
}
