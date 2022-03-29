package util

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
)

// Refer to https://github.com/golang/go/issues/28804
func IsRunAsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get SID")
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	isAdminMember, err := token.IsMember(sid)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get token membership")
		return false
	}

	return token.IsElevated() || isAdminMember
}

func GetAdminUserName() string {
	return "administrator"
}
