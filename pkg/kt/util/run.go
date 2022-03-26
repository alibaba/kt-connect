package util

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// CMDContext context of cmd
type CMDContext struct {
	Ctx     context.Context
	Cmd     *exec.Cmd
	Name    string
	Stop    chan struct{} // notify parent current Cmd occur error
}

// RunAndWait run cmd
func RunAndWait(cmd *exec.Cmd) (string, string, error) {
	outbuf, errbuf, err := runCmd(&CMDContext{
		Cmd:     cmd,
		Name:    cmd.Path,
	})
	if err != nil {
		return outbuf.String(), errbuf.String(), err
	}
	err = cmd.Wait()
	return outbuf.String(), errbuf.String(), err
}

// BackgroundRun run cmd in background with context
func BackgroundRun(cmdCtx *CMDContext) error {
	outbuf, errbuf, err := runCmd(cmdCtx)
	if err != nil {
		return err
	}
	go func() {
		err = cmdCtx.Cmd.Wait()
		stdout := strings.TrimSpace(outbuf.String())
		stderr := strings.TrimSpace(errbuf.String())
		if len(stdout) > 0 {
			log.Debug().Msgf("[STDOUT] %s", stdout)
		}
		if len(stderr) > 0 {
			log.Debug().Msgf("[STDERR] %s", stderr)
		}
		if err != nil {
			log.Info().Msgf("Background task %s closed, %s", cmdCtx.Name, err.Error())
		}
		log.Info().Msgf("Task %s completed", cmdCtx.Name)
	}()
	return nil
}

// CanRun check whether a command can execute successful
func CanRun(cmd *exec.Cmd) bool {
	return cmd.Run() == nil
}

func runCmd(cmdCtx *CMDContext) (*bytes.Buffer, *bytes.Buffer, error) {
	cmd := cmdCtx.Cmd
	log.Debug().Msgf("Task name = %s, cmd.Args = %+v", cmdCtx.Name, cmd.Args)

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
	log.Debug().Msgf("Start %s at pid: %d", cmdCtx.Name, pid)
	// will kill the process when parent cancel
	go func() {
		if cmdCtx.Ctx != nil {
			select {
			case <-cmdCtx.Ctx.Done():
				err2 := cmd.Process.Kill()
				if err2 != nil {
					log.Debug().Msgf("Task %s(%d) killed", cmdCtx.Name, pid)
				}
			}
		}
	}()
	return outbuf, errbuf, nil
}
