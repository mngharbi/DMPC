package daemon

/*
	Utilities/Definitions for config
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/pipeline"
	"github.com/mngharbi/DMPC/status"
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
type NumWorkersOnlyConfig struct {
	NumWorkers int `json:"numWorkers"`
}
type Config struct {
	// Log level setting
	LogLevel core.LogLevel `json:"logLevel"`

	// All customizable paths
	Paths ConfigPaths `json:"paths"`

	// Configuration for users subsystem
	Users NumWorkersOnlyConfig `json:"users"`

	// Configuration for users subsystem
	Status StatusSubsystemConfig `json:"status"`

	// Configuration for executor subsystem
	Executor NumWorkersOnlyConfig `json:"executor"`

	// Configuration for decryptor subsystem
	Decryptor NumWorkersOnlyConfig `json:"decryptor"`

	// Configuration for pipeline subsystem (websocket)
	Pipeline PipelineSubsystemConfig `json:"pipeline"`
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

func (config *Config) getUsersSubsystemConfig() users.Config {
	return users.Config{
		NumWorkers: config.Users.NumWorkers,
	}
}

type StatusSubsystemConfig struct {
	Update    NumWorkersOnlyConfig `json:"update"`
	Listeners NumWorkersOnlyConfig `json:"listeners"`
}

func (config *Config) getStatusSubsystemConfig() (status.StatusServerConfig, status.ListenersServerConfig) {
	return status.StatusServerConfig{
			NumWorkers: config.Status.Update.NumWorkers,
		}, status.ListenersServerConfig{
			NumWorkers: config.Status.Listeners.NumWorkers,
		}
}

func (config *Config) getExecutorSubsystemConfig() executor.Config {
	return executor.Config{
		NumWorkers: config.Executor.NumWorkers,
	}
}

func (config *Config) getDecryptorSubsystemConfig() decryptor.Config {
	return decryptor.Config{
		NumWorkers: config.Decryptor.NumWorkers,
	}
}

type PipelineSubsystemConfig struct {
	CheckOrigin bool   `json:"checkOrigin"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
}

func (config *Config) getPipelineSubsystemConfig() pipeline.Config {
	return pipeline.Config{
		CheckOrigin: config.Pipeline.CheckOrigin,
		Hostname:    config.Pipeline.Hostname,
		Port:        config.Pipeline.Port,
	}
}
