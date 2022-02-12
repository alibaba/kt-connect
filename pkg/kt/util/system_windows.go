package util

import (
	"golang.org/x/sys/windows"
)

func IsRunAsAdmin() bool {
	// Refer to telepresence, which is transform from windows official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	adm, err := windows.GetCurrentProcessToken().IsMember(sid)
	return err == nil && adm
}

func GetAdminUserName() string {
	return "administrator"
}
