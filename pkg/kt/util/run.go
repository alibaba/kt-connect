package util

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	outbuf, errbuf, err := runCmd(cmd, cmd.Path)
	if err != nil {
		return outbuf.String(), errbuf.String(), err
	}
	err = cmd.Wait()
	return outbuf.String(), errbuf.String(), err
}

// BackgroundRun run cmd in background with context
func BackgroundRun(cmd *exec.Cmd, name string, res chan error) error {
	outbuf, errbuf, err := runCmd(cmd, name)
	if err != nil {
		return err
	}

	go func() {
		err = cmd.Wait()
		stdout := strings.TrimSpace(outbuf.String())
		stderr := strings.TrimSpace(errbuf.String())
		if len(stdout) > 0 {
			log.Debug().Msgf("[STDOUT] %s", stdout)
		}
		if len(stderr) > 0 {
			log.Debug().Msgf("[STDERR] %s", stderr)
		}
		if err != nil {
			log.Info().Msgf("Background task %s closed, %s", name, err.Error())
		} else {
			log.Info().Msgf("Background task %s completed", name)
		}
		res <-err
	}()

	return nil
}

// CanRun check whether a command can execute successful
func CanRun(cmd *exec.Cmd) bool {
	return cmd.Run() == nil
}

func runCmd(cmd *exec.Cmd, name string) (*bytes.Buffer, *bytes.Buffer, error) {
	log.Debug().Msgf("Task name = %s, cmd.Args = %+v", name, cmd.Args)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	if stdout != nil && stderr != nil {
		go io.Copy(outbuf, stdout)
		go io.Copy(errbuf, stderr)
	}

	if err := cmd.Start(); err != nil {
		return outbuf, errbuf, err
	}

	time.Sleep(100 * time.Millisecond)
	pid := cmd.Process.Pid
	log.Debug().Msgf("Start %s at pid: %d", name, pid)
	return outbuf, errbuf, nil
}
