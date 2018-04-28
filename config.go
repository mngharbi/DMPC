package main

/*
	Utilities/Definitions for config
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
	"io/ioutil"
	"os/user"
)

/*
	Structure for config file
*/
type ConfigPaths struct {
	// Path to file containing base root user object
	RootUserFilePath string `json:"userFile"`

	// Key paths
	PublicEncryptionKeyPath  string `json:"publicEncryptionKeyPath"`
	PrivateEncryptionKeyPath string `json:"privateEncryptionKeyPath"`
	PublicSigningKeyPath     string `json:"publicSigningKeyPath"`
	PrivateSigningKeyPath    string `json:"privateSigningKeyPath"`
}
type Config struct {
	// Log level setting
	LogLevel core.LogLevel `json:"logLevel"`

	// All customizable paths
	Paths ConfigPaths `json:"paths"`

	// Configuration for users subsystem
	Users UserSubsystemConfig `json:"users"`
}

/*
	Constants for config directories and files
*/
const (
	dmpcDir        string = ".dmpc"
	setupFilename  string = ".setup"
	configFilename string = "config.json"
)

/*
	Error messages
*/
const (
	configurationNotFound string = "Configuration file not found"
)

/*
	Utilities for getting directories
*/
func getRootDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return usr.HomeDir + "/"
}

func getDmpcDir() string {
	return getRootDir() + dmpcDir + "/"
}

func getSetupFilename() string {
	return getDmpcDir() + setupFilename
}

func getConfigFilename() string {
	return getDmpcDir() + configFilename
}

/*
	Get config structure from config file
*/
func getConfig() *Config {
	raw, err := ioutil.ReadFile(getConfigFilename())
	if err != nil {
		log.Fatalf(configurationNotFound)
		return nil
	}

	var config Config
	json.Unmarshal(raw, &config)
	return &config
}

/*
	Server configuration
*/
type UserSubsystemConfig struct {
	NumWorkers int `json:"numWorkers"`
}

func (config *Config) getUsersSubsystemConfig() users.Config {
	return users.Config{
		NumWorkers: config.Users.NumWorkers,
	}
}
