package clean

import (
	"testing"
)

func Test_toPid(t *testing.T) {
	component, pid := parseComponentAndPid("connect-123.pid")
	if "connect" != component || 123 != pid {
		t.Errorf("unmatch %d", pid)
	}
	component, pid = parseComponentAndPid("connect-abc.pid")
	if "" != component || -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
	component, pid = parseComponentAndPid("abc")
	if "" != component || -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
}
