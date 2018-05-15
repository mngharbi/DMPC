/*
	Common utilities for file I/O
*/

package startup

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

/*
	Utilities for building paths
*/
func getRootDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return usr.HomeDir + "/"
}

/*
	Tests existence
*/
func PathExists(paths ...string) bool {
	_, err := os.Stat(GetInstallPath(paths...))
	return err == nil
}

/*
	Reads file
*/
func ReadFile(paths ...string) ([]byte, error) {
	return ioutil.ReadFile(GetInstallPath(paths...))
}

/*
	Write to file (creates if file doesn't exist)
*/
func WriteFile(data []byte, paths ...string) error {
	return ioutil.WriteFile(GetInstallPath(paths...), data, os.ModePerm)
}

/*
	Makes directory
*/
func MkdirAll(paths ...string) {
	os.MkdirAll(GetInstallPath(paths...), os.ModePerm)
}
