package integration

import "flag"

var timeOutMultiplier = flag.Float64("timeout-multiplier", 1, "multiply the timeout for the tests")
