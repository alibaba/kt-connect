package process

// Interrupt ...
func Interrupt() chan struct{} {
	return make(chan struct{})
}
