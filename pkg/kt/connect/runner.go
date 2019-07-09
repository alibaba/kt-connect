package connect

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string, debug bool) (err error) {

	if debug {
		log.Printf("Child, os.Args = %+v\n", os.Args)
		log.Printf("Child, cmd.Args = %+v\n", cmd.Args)
	}

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

	go func() {
		err = cmd.Wait()
		log.Printf("%s exited\n", name)
	}()

	time.Sleep(time.Duration(2) * time.Second)
	pid := cmd.Process.Pid
	log.Printf("%s start at pid: %d\n", name, pid)
	return
}
