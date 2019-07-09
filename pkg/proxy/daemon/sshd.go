package daemon

import (
	"fmt"
	"os/exec"
)

func SSHDStart() {
	cmd := exec.Command("/usr/sbin/sshd", "-D")
	err := cmd.Start()
	if err != nil {
		return
	}
	pid := cmd.Process.Pid
	fmt.Printf("SSHD start at pid: %d\n", pid)
	go func() {
		err = cmd.Wait()
		fmt.Printf("SSHD Exited with error: %v\n", err)
	}()
}
