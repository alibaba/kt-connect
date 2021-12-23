package connect

import "testing"

func TestAllocateIP(t *testing.T) {
	cidr := "10.1.1.0/30"

	srcIP, destIP, err := allocateTunIP(cidr)
	if err != nil {
		t.Errorf("allocateTunIP() error: %v", err)
	}

	if srcIP != "10.1.1.1" {
		t.Errorf("allocateTunIP() failed, current: %s, want: %s", srcIP, "10.1.1.1")
	}

	if destIP != "10.1.1.2" {
		t.Errorf("allocateTunIP() failed, current: %s, want: %s", destIP, "10.1.1.2")
	}
}
