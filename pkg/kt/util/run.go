package util

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"os/exec"
)

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	log.Debug().Msgf("Task %s with args %+v", cmd.Path, cmd.Args)
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// BackgroundRun run cmd in background with context
func BackgroundRun(cmd *exec.Cmd, name string, res chan error) error {
	cmd.Stderr = BackgroundLogger
	cmd.Stdout = BackgroundLogger
	log.Debug().Msgf("Task %s with args %+v", name, cmd.Args)
	if err := cmd.Start(); err != nil {
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
