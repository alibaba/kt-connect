package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// CMDContext ...
type CMDContext struct {
	Ctx  context.Context
	Cmd  *exec.Cmd
	Name string
	Stop chan bool // notify parent current Cmd occur error
}

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd, name string, debug bool) (err error) {
	ctx := &CMDContext{
		Cmd:  cmd,
		Name: name,
	}
	err = runCmd(ctx, debug)
	if err != nil {
		return
	}
	err = cmd.Wait()
	return
}

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string, debug bool) (err error) {
	ctx := &CMDContext{
		Cmd:  cmd,
		Name: name,
	}
	err = runCmd(ctx, debug)
	if err != nil {
		return
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			return
		}
		log.Info().Msgf("%s finished", name)
	}()
	return
}

// BackgroundRunWithCtx run cmd in backgroud with context
func BackgroundRunWithCtx(cmdCtx *CMDContext, debug bool) (err error) {
	err = runCmd(cmdCtx, debug)
	if err != nil {
		return
	}
	go func() {
		if err = cmdCtx.Cmd.Wait(); err != nil {
			if !strings.Contains(err.Error(), "signal:") {
				log.Error().Err(err).Msgf("background process of %s failed", cmdCtx.Name)
			}
			cmdCtx.Stop <- true
			return
		}
		log.Info().Msgf("%s finished with context", cmdCtx.Name)
	}()
	return
}

func runCmd(cmdCtx *CMDContext, debug bool) (err error) {
	cmd := cmdCtx.Cmd
	log.Debug().Msgf("Child, os.Args = %+v", os.Args)
	log.Debug().Msgf("Child, name = %s, cmd.Args = %+v", cmdCtx.Name, cmd.Args)

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
		cmdCtx.Stop <- true
		return
	}

	time.Sleep(time.Duration(1) * time.Second)
	pid := cmd.Process.Pid
	log.Info().Msgf("%s start at pid: %d", cmdCtx.Name, pid)
	// will kill the process when parent cancel
	go func() {
		if cmdCtx.Ctx != nil {
			select {
			case <-cmdCtx.Ctx.Done():
				err := cmd.Process.Kill()
				if err != nil {
					log.Error().Msgf(err.Error())
				}
			}
		}
	}()
	return
}
