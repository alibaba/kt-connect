package dnsserver

type DomainNotExistError struct {
	name string
}

func (e DomainNotExistError) Error() string {
	return "domain " + e.name + " not exist"
}

func IsDomainNotExist(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(DomainNotExistError)
	return ok
}
