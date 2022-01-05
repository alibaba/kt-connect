package mesh

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_toPortMapParameter(t *testing.T) {
	require.Equal(t, toPortMapParameter(map[int]int{ }), "", "port map parameter incorrect")
	require.Equal(t, toPortMapParameter(map[int]int{ 80:8080 }), "80:8080", "port map parameter incorrect")
	res := toPortMapParameter(map[int]int{ 80:8080, 70:7000 })
	require.True(t, res == "80:8080,70:7000" || res == "70:7000,80:8080", "port map parameter incorrect")
}
