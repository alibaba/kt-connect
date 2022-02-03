package command

import (
	"testing"
)

func Test_toPid(t *testing.T) {
	action := Action{}
	pid := action.toPid("connect-123.pid")
	if 123 != pid {
		t.Errorf("unmatch %d", pid)
	}
	pid = action.toPid("connect-abc.pid")
	if -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
	pid = action.toPid("abc")
	if -1 != pid {
		t.Errorf("unmatch %d", pid)
	}
}
