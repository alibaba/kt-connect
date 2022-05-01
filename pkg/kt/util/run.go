package util

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"io"
	"os/exec"
)

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	outBuf := bytes.NewBufferString("")
	errBuf := bytes.NewBufferString("")
	if stdout != nil && stderr != nil {
		go io.Copy(outBuf, stdout)
		go io.Copy(errBuf, stderr)
	}

	if err := runCmd(cmd, cmd.Path); err != nil {
		return outBuf.String(), errBuf.String(), err
	}
	return outBuf.String(), errBuf.String(), cmd.Wait()
}

// BackgroundRun run cmd in background with context
func BackgroundRun(cmd *exec.Cmd, name string, res chan error) error {
	cmd.Stderr = BackgroundLogger
	cmd.Stdout = BackgroundLogger
	if err := runCmd(cmd, name); err != nil {
		return err
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Debug().Msgf("Background task %s closed, %s", name, err.Error())
		} else {
			log.Debug().Msgf("Background task %s completed", name)
		}
		res <-err
	}()

	return nil
}

// CanRun check whether a command can execute successful
func CanRun(cmd *exec.Cmd) bool {
	return cmd.Run() == nil
}

func runCmd(cmd *exec.Cmd, name string) error {
	log.Debug().Msgf("Task %s with args %+v", name, cmd.Args)
	return cmd.Start()
}
