package cli

import (
	"log"
	"os"
	"path"
)

/*
	Utilities for getting paths
*/
func GetInstallPath(paths ...string) string {
	full := []string{getRootDir(), ConfigDir}
	full = append(full, paths...)
	return path.Join(full...)
}

func GetInstallRoot() string {
	return GetInstallPath()
}

/*
	Creates bad state file
*/
func MakeBadStateFile() {
	if err := WriteFile([]byte{}, BadStateFilename); err != nil {
		log.Fatalf("Failed to create bad state file. err=%v", err)
	}
}

/*
	Deletes base directory
*/
func DeleteDmpcDir() {
	os.RemoveAll(GetInstallRoot())
}

/*
	Makes base directory
*/
func MakeDmpcDir() {
	MkdirAll()
}

/*
	Checking install state
*/
func IsFunctional() bool {
	return IsInstalled() && !IsInBadState()
}

func IsInstalled() bool {
	return PathExists(ConfigFilename)
}

func IsInBadState() bool {
	return PathExists(BadStateFilename)
}
