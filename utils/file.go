package utils

import "os"

func IsFileExist(fileName string) bool {
	_, err := os.Stat(fileName)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return true
}
