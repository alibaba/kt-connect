package util

func init() {
	CreateDirIfNotExist(KtHome)
	FixFileOwner(KtHome)
}
