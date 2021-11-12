package general

import (
	"github.com/urfave/cli"
)

// NewCliAuthor return cli author
func NewCliAuthor() []cli.Author {
	return []cli.Author{
		{
			Name: "Alibaba Cloud",
		},
	}
}
