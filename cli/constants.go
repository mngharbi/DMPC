package cli

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
)

/*
	Constants for default config directories and files
*/
const (
	ConfigDir             string = ".dmpc"
	BadStateFilename      string = ".badstate"
	ConfigFilename        string = "config.json"
	RootUserFilename      string = "user.json"
	KeysDir               string = "keys"
	EncryptionKeyFilename string = "encryption_rsa"
	SigningKeyFilename    string = "signing_rsa"
	PublicKeySuffix       string = ".pub"
)

/*
	Default root user object
*/
var defaultUserObject users.UserObject = users.UserObject{
	Permissions: users.PermissionsObject{
		Channel: users.ChannelPermissionsObject{
			Add:  true,
			Read: true,
		},
		User: users.UserPermissionsObject{
			Add:               true,
			Read:              true,
			Remove:            true,
			EncKeyUpdate:      true,
			SignKeyUpdate:     true,
			PermissionsUpdate: true,
		},
	},
	Active: true,
}

/*
	Default daemon config
*/
var defaultDaemonConfig Config = Config{
	LogLevel: core.INFO,
	Locker: NumWorkersOnlyConfig{
		NumWorkers: 4,
	},
	Users: NumWorkersOnlyConfig{
		NumWorkers: 4,
	},
	Channels: ChannelsSubsystemConfig{
		Channels: NumWorkersOnlyConfig{
			NumWorkers: 2,
		},
		Messages: NumWorkersOnlyConfig{
			NumWorkers: 4,
		},
		Listeners: NumWorkersOnlyConfig{
			NumWorkers: 2,
		},
	},
	Status: StatusSubsystemConfig{
		Update: NumWorkersOnlyConfig{
			NumWorkers: 2,
		},
		Listeners: NumWorkersOnlyConfig{
			NumWorkers: 4,
		},
	},
	Keys: NumWorkersOnlyConfig{
		NumWorkers: 4,
	},
	Executor: NumWorkersOnlyConfig{
		NumWorkers: 4,
	},
	Decryptor: NumWorkersOnlyConfig{
		NumWorkers: 4,
	},
	Pipeline: PipelineSubsystemConfig{
		CheckOrigin: false,
		Port:        64927,
	},
}
