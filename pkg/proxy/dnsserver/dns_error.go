package dnsserver

// DomainNotExistError ...
type DomainNotExistError struct {
	name string
}

func (e DomainNotExistError) Error() string {
	return "domain " + e.name + " not exist"
}

// IsDomainNotExist check the error type
func IsDomainNotExist(err error) bool {
	_, ok := err.(DomainNotExistError)
	return ok
}
