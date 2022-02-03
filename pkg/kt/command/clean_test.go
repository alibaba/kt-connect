package command

import (
	"testing"
)

func Test_toPid(t *testing.T) {
	action := Action{}
	component, pid := action.parseComponentAndPid("connect-123.pid")
	if "connect" != component || 123 != pid {
		t.Errorf("unmatch %d", pid)
	}
	component, pid = action.parseComponentAndPid("connect-abc.pid")
	if "" != component || -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
	component, pid = action.parseComponentAndPid("abc")
	if "connect" != component || -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
}
