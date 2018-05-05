package config

/*
	Utilities/Definitions for conf
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/pipeline"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"log"
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
	Error messages
*/
const (
	configurationNotFound      string = "Configuration file not found"
	invalidConfigurationFormat string = "Invalid configuration file format"
)

/*
	(En/De)coding
*/
func (conf *Config) Encode() ([]byte, error) {
	return json.Marshal(conf)
}

func (conf *Config) Decode(encoded []byte) error {
	return json.Unmarshal(encoded, conf)
}

/*
	Get config structure from config file
*/
func GetConfig() *Config {
	raw, err := ReadFile(ConfigFilename)
	if err != nil {
		log.Fatalf(configurationNotFound)
		return nil
	}
	var conf Config
	if err := conf.Decode(raw); err != nil {
		log.Fatalf(invalidConfigurationFormat)
		return nil
	}
	return &conf
}

/*
	Server confuration
*/

func (conf *Config) GetUsersSubsystemConfig() users.Config {
	return users.Config{
		NumWorkers: conf.Users.NumWorkers,
	}
}

type StatusSubsystemConfig struct {
	Update    NumWorkersOnlyConfig `json:"update"`
	Listeners NumWorkersOnlyConfig `json:"listeners"`
}

func (conf *Config) GetStatusSubsystemConfig() (status.StatusServerConfig, status.ListenersServerConfig) {
	return status.StatusServerConfig{
			NumWorkers: conf.Status.Update.NumWorkers,
		}, status.ListenersServerConfig{
			NumWorkers: conf.Status.Listeners.NumWorkers,
		}
}

func (conf *Config) GetExecutorSubsystemConfig() executor.Config {
	return executor.Config{
		NumWorkers: conf.Executor.NumWorkers,
	}
}

func (conf *Config) GetDecryptorSubsystemConfig() decryptor.Config {
	return decryptor.Config{
		NumWorkers: conf.Decryptor.NumWorkers,
	}
}

type PipelineSubsystemConfig struct {
	CheckOrigin bool   `json:"checkOrigin"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
}

func (conf *Config) GetPipelineSubsystemConfig() pipeline.Config {
	return pipeline.Config{
		CheckOrigin: conf.Pipeline.CheckOrigin,
		Hostname:    conf.Pipeline.Hostname,
		Port:        conf.Pipeline.Port,
	}
}
