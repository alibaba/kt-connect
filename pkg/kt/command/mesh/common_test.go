package mesh

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_isValidKey(t *testing.T) {
	validCases := []string{"kt", "k123", "kt-version_1123"}
	for _, c := range validCases {
		require.True(t, isValidKey(c), "'%s' should be valid", c)
	}
	invalidCases := []string{"-version_1123", "123", "", "kt.ver"}
	for _, c := range invalidCases {
		require.True(t, !isValidKey(c), "'%s' should be invalid", c)
	}
}

func Test_getVersion(t *testing.T) {
	var k, v string
	k, v = getVersion("")
	require.Equal(t, k, "version")
	require.Equal(t, len(v), 5)
	k, v = getVersion("test")
	require.Equal(t, k, "version")
	require.Equal(t, v, "test")
	k, v = getVersion("mark:")
	require.Equal(t, k, "mark")
	require.Equal(t, len(v), 5)
	k, v = getVersion("mark:test")
	require.Equal(t, k, "mark")
	require.Equal(t, v, "test")
}
