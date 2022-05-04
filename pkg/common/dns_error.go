package common

import "fmt"

// DomainNotExistError ...
type DomainNotExistError struct {
	name string
	qtype uint16
}

func (e DomainNotExistError) Error() string {
	return fmt.Sprintf("domain %s (%d) not exist", e.name, e.qtype)
}

// IsDomainNotExist check the error type
func IsDomainNotExist(err error) bool {
	_, exists := err.(DomainNotExistError)
	return exists
}
