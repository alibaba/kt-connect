package process

var interrupt = make(chan struct{})

// Stop ...
func Stop(stop struct{}, cancel func()) {
	if cancel == nil {
		return
	}
	cancel()
	interrupt <- stop
}

// Interrupt ...
func Interrupt() chan struct{} {
	return interrupt
}
