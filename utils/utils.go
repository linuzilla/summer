package utils

import (
	"path/filepath"
	"strings"
)

func Basename(fileName string) string {
	if i := strings.LastIndex(fileName, "/"); i == -1 {
		return fileName
	} else {
		return fileName[i+1:]
	}
}

func FileNameToExportedVariable(fileName string) string {
	return strings.Replace(
		strings.Title(
			strings.Replace(
				strings.TrimSuffix(fileName, filepath.Ext(fileName)),
				"_", " ", -1)),
		" ", "", -1)
}

func SetterName(variableName string) string {
	return `Set` + strings.Title(variableName)
}
