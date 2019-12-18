package util

// Labels migration labels
func Labels(name string, component string, references map[string]string, extraLabels string) (labels map[string]string) {
	labels = map[string]string{
		"kt":           name,
		"kt-component": component,
		"control-by":   "kt",
	}

	for k, v := range String2Map(extraLabels) {
		labels[k] = v
	}

	for k, v := range references {
		labels[k] = v
	}

	return
}
