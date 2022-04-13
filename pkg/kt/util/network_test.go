package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractHostIp(t *testing.T) {
	require.Equal(t, "1.2.3.4", ExtractHostIp("http://1.2.3.4"))
	require.Equal(t, "1.2.3.4", ExtractHostIp("http://1.2.3.4:8080"))
	require.Equal(t, "1.2.3.4", ExtractHostIp("http://1.2.3.4:8080/a/b/c"))
	require.Equal(t, "127.0.0.1", ExtractHostIp("http://localhost:8080/a/b/c"))
}
