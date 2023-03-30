package config

import (
	"fmt"
	"testing"
)

func TestSaveProfile(t *testing.T) {
	err := SaveProfile([]string{"test"})
	fmt.Println(err)
}
