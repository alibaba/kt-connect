package hack

import (
	_ "embed"
)

//go:embed kube/config
var CustomizeKubeConfig string

//go:embed kt/config
var CustomizeKtConfig string
