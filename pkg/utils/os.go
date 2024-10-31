package utils

import "os"

func ExistFile(filePath string) bool {
	if info, err := os.Stat(filePath); err == nil {
		if !info.IsDir() {
			return true
		}
	}
	return false
}

func ExistDir(dirPath string) bool {
	if info, err := os.Stat(dirPath); err == nil {
		if info.IsDir() {
			return true
		}
	}
	return false
}
