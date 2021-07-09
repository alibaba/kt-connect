package common

// CommandExecError ...
type CommandExecError struct {
	Reason string
}

func (e CommandExecError) Error() string {
	return "command execution failed with " + e.Reason
}

// IsCommandExecError check the error type
func IsCommandExecError(err error) bool {
	_, ok := err.(CommandExecError)
	return ok
}
