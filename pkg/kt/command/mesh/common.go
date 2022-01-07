package mesh

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

func getVersion(versionMark string) (string, string) {
	versionKey := "version"
	versionVal := strings.ToLower(util.RandomString(5))
	if len(versionMark) != 0 {
		versionParts := strings.Split(versionMark, ":")
		if len(versionParts) > 1 {
			if isValidKey(versionParts[0]) {
				versionKey = versionParts[0]
			} else {
				log.Warn().Msgf("mark key '%s' is invalid, using default key '%s'", versionParts[0], versionKey)
			}
			if len(versionParts[1]) > 0 {
				versionVal = versionParts[1]
			}
		} else {
			versionVal = versionParts[0]
		}
	}
	return versionKey, versionVal
}

func isValidKey(key string) bool {
	ok, err := regexp.MatchString("^[a-z][a-z0-9_-]*$", key)
	return err == nil && ok
}
