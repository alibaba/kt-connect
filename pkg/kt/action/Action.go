package action

// Action cmd action
type Action struct {
	Kubeconfig string
	Namespace  string
	Debug      bool
	Image      string
	PidFile    string
	UserHome   string
}
