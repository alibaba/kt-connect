package util

func init() {
	_ = CreateDirIfNotExist(KtHome)
	FixFileOwner(KtHome)
}
