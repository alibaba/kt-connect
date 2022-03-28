package dns

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDropHosts(t *testing.T) {
	type args struct {
		linesBeforeDrop []string
		linesAfterDrop  []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty hosts",
			args: args{
				linesBeforeDrop: []string{},
				linesAfterDrop:  []string{},
			},
		},
		{
			name: "no kt hosts to drop",
			args: args{
				linesBeforeDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
				linesAfterDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
			},
		},
		{
			name: "kt hosts at beginning",
			args: args{
				linesBeforeDrop: []string{
					"# Kt Hosts Begin",
					"172.12.3.4 tomcat",
					"# Kt Hosts End",
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
				linesAfterDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
			},
		},
		{
			name: "kt hosts at the end",
			args: args{
				linesBeforeDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
					"# Kt Hosts Begin",
					"172.12.3.4 tomcat",
					"# Kt Hosts End",
				},
				linesAfterDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
			},
		},
		{
			name: "kt hosts at middle",
			args: args{
				linesBeforeDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"# Kt Hosts Begin",
					"172.12.3.4 tomcat",
					"# Kt Hosts End",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
				linesAfterDrop: []string{
					"# Local",
					"127.0.0.1 localhost",
					"::1 localhost",
					"",
					"# Added by Docker Desktop",
					"127.0.0.1 kubernetes.docker.internal",
					"",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linesAfterDrop, _, _ := dropHosts(tt.args.linesBeforeDrop, "")
			require.Equal(t, len(tt.args.linesAfterDrop), len(linesAfterDrop),
				"should has %d lines, but got %d", len(tt.args.linesAfterDrop), len(linesAfterDrop))
			for i, line := range tt.args.linesAfterDrop {
				require.Equal(t, line, linesAfterDrop[i],
					"hosts line %d mismatch: expect [%s] got [%s]", i, line, linesAfterDrop[i])
			}
		})
	}
}

func TestDumpHosts(t *testing.T) {
	type args struct {
		hostsToDump    map[string]string
		linesAfterDump []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty hosts",
			args: args{
				hostsToDump:    map[string]string{},
				linesAfterDump: []string{"# Kt Hosts Begin", "# Kt Hosts End"},
			},
		},
		{
			name: "many hosts",
			args: args{
				hostsToDump:    map[string]string{"tomcat": "192.12.3.4", "nginx": "192.12.5.6"},
				linesAfterDump: []string{"# Kt Hosts Begin", "192.12.3.4 tomcat", "192.12.5.6 nginx", "# Kt Hosts End"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linesAfterDump := dumpHosts(tt.args.hostsToDump, []string{})
			require.Equal(t, len(tt.args.linesAfterDump), len(linesAfterDump),
				"should has %d lines, but got %d", len(tt.args.linesAfterDump), len(linesAfterDump))
			for _, line := range tt.args.linesAfterDump {
				require.Contains(t, linesAfterDump, line, "hosts %s missing", line)
			}
		})
	}
}

func TestMergeHost(t *testing.T) {
	type args struct {
		linesBegin      []string
		linesEnd        []string
		linesAfterMerge []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "emtpy merge",
			args: args{
				linesBegin:      []string{},
				linesEnd:        []string{},
				linesAfterMerge: []string{"", ""},
			},
		},
		{
			name: "emtpy lines begin",
			args: args{
				linesBegin:      []string{},
				linesEnd:        []string{"abc", "def"},
				linesAfterMerge: []string{"", "abc", "def", ""},
			},
		},
		{
			name: "emtpy lines end",
			args: args{
				linesBegin:      []string{"abc", "def"},
				linesEnd:        []string{},
				linesAfterMerge: []string{"abc", "def", "", ""},
			},
		},
		{
			name: "common merge",
			args: args{
				linesBegin:      []string{"abc", "def"},
				linesEnd:        []string{"ghi", "lmn"},
				linesAfterMerge: []string{"abc", "def", "", "ghi", "lmn", ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linesAfterMerge := mergeLines(tt.args.linesBegin, tt.args.linesEnd)
			require.Equal(t, len(tt.args.linesAfterMerge), len(linesAfterMerge),
				"should has %d lines, but got %d", len(tt.args.linesAfterMerge), len(linesAfterMerge))
			for i, line := range tt.args.linesAfterMerge {
				require.Equal(t, line, linesAfterMerge[i],
					"hosts line %d mismatch: expect [%s] got [%s]", i, line, linesAfterMerge[i])
			}
		})
	}
}
