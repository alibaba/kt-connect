package mesh

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

func createShadowAndInbound(ctx context.Context, k cluster.KubernetesInterface, shadowPodName string,
	labels, annotations map[string]string, options *options.DaemonOptions) error {

	labels[common.ControlBy] = common.KubernetesTool
	envs := make(map[string]string)
	_, podName, _, err := cluster.GetOrCreateShadow(ctx, k, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}

	// record context data
	options.RuntimeOptions.Shadow = shadowPodName

	shadow := connect.Create(options)
	if _, err = shadow.Inbound(options.MeshOptions.Expose, podName); err != nil {
		return err
	}
	return nil
}

func getVersion(versionMark string) (string, string) {
	versionKey := "kt-version"
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
