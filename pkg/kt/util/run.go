package util

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// CMDContext context of cmd
type CMDContext struct {
	Ctx     context.Context
	Cmd     *exec.Cmd
	Name    string
	IsDebug bool
	Stop    chan struct{} // notify parent current Cmd occur error
}

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd, isDebug bool) error {
	ctx := &CMDContext{
		Cmd:     cmd,
		Name:    cmd.Path,
		IsDebug: isDebug,
	}
	if err := runCmd(ctx); err != nil {
		return err
	}
	return cmd.Wait()
}

// BackgroundRun run cmd in background with context
func BackgroundRun(cmdCtx *CMDContext) error {
	if err := runCmd(cmdCtx); err != nil {
		return err
	}
	go func() {
		if err := cmdCtx.Cmd.Wait(); err != nil {
			log.Info().Msgf("Background process %s exit abnormally: %s", cmdCtx.Name, err.Error())
		}
		log.Info().Msgf("Finished %s with context", cmdCtx.Name)
	}()
	return nil
}

// CanRun check whether a command can execute successful
func CanRun(cmd *exec.Cmd) bool {
	return cmd.Run() == nil
}

func runCmd(cmdCtx *CMDContext) error {
	var err error
	cmd := cmdCtx.Cmd
	log.Debug().Msgf("Child, name = %s, cmd.Args = %+v", cmdCtx.Name, cmd.Args)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if cmdCtx.IsDebug && stdout != nil && stderr != nil {
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
	pid := cmd.Process.Pid
	log.Debug().Msgf("Start %s at pid: %d", cmdCtx.Name, pid)
	// will kill the process when parent cancel
	go func() {
		if cmdCtx.Ctx != nil {
			select {
			case <-cmdCtx.Ctx.Done():
				err2 := cmd.Process.Kill()
				if err2 != nil {
					log.Debug().Msgf("Process %d competed", pid)
				}
			}
		}
	}()
	return nil
}
