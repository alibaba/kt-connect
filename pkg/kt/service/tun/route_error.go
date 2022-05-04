package tun

import "fmt"

// AllRouteFailError ...
type AllRouteFailError struct {
	originalError error
}

func (e AllRouteFailError) Error() string {
	return fmt.Sprintf("all routes failed to setup")
}

func (e AllRouteFailError) OriginalError() error {
	return e.originalError
}

// IsAllRouteFailError check the error type
func IsAllRouteFailError(err error) bool {
	_, exists := err.(AllRouteFailError)
	return exists
}
