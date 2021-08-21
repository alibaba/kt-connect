package exec

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
	Ctx  context.Context
	Cmd  *exec.Cmd
	Name string
	Stop chan struct{} // notify parent current Cmd occur error
}

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd, name string) (err error) {
	ctx := &CMDContext{
		Cmd:  cmd,
		Name: name,
	}
	err = runCmd(ctx)

	if err != nil {
		return
	}
	err = cmd.Wait()
	return
}

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string) (err error) {
	ctx := &CMDContext{
		Cmd:  cmd,
		Name: name,
	}
	err = runCmd(ctx)
	if err != nil {
		return
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			return
		}
		log.Info().Msgf("Finished %s", name)
	}()
	return
}

// BackgroundRunWithCtx run cmd in background with context
func BackgroundRunWithCtx(cmdCtx *CMDContext) (err error) {
	err = runCmd(cmdCtx)
	if err != nil {
		return
	}
	go func() {
		if err = cmdCtx.Cmd.Wait(); err != nil {
			log.Info().Msgf("Background process %s exit abnormally: %s", cmdCtx.Name, err.Error())
		}
		log.Info().Msgf("Finished %s with context", cmdCtx.Name)
	}()
	return
}

func runCmd(cmdCtx *CMDContext) error {
	var err error
	cmd := cmdCtx.Cmd
	log.Debug().Msgf("Child, os.Args = %+v", os.Args)
	log.Debug().Msgf("Child, name = %s, cmd.Args = %+v", cmdCtx.Name, cmd.Args)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if stdout != nil && stderr != nil {
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
	}

	err = cmd.Start()
	if err != nil {
		cmdCtx.Stop <- struct{}{}
		return err
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
	pid := cmd.Process.Pid
	log.Info().Msgf("Start %s at pid: %d", cmdCtx.Name, pid)
	// will kill the process when parent cancel
	go func() {
		if cmdCtx.Ctx != nil {
			select {
			case <-cmdCtx.Ctx.Done():
				err := cmd.Process.Kill()
				if err != nil {
					log.Debug().Msgf("Process %d already competed", pid)
				}
			}
		}
	}()
	return err
}
