package general

func UsageTemplate(showInheritedFlags bool) string {
	usage := `{{if .HasExample}}Usage:
{{.Example}}

{{end}}{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

{{end}}{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}`

	if showInheritedFlags {
		usage += `
	
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}`
	}

	usage += `

Use "{{.CommandPath}} [command] --help" for more information about a command.
`
	return usage
}
