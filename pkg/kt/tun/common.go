package tun

import (
	"net"
	"strconv"
	"strings"
)

func toIpAndMask(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	val := make([]byte, len(ipNet.Mask))
	copy(val, ipNet.Mask)

	var s []string
	for _, i := range val[:] {
		s = append(s, strconv.Itoa(int(i)))
	}
	return ipNet.IP.String(), strings.Join(s, "."), nil
}
