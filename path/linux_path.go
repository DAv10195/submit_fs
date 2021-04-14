// +build !windows

package path

func GetDefaultConfigFilePath() string {
	return "/etc/submit-file-server/"
}

func GetDefaultWorkDirPath() string {
	return "var/cache/submit-file-server"
}
