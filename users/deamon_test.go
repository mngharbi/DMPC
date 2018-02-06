package users

import (
	"testing"
)

func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
	}
}

func singleWorkerConfig() Config {
	return Config{
		NumWorkers: 1,
	}
}

func TestStartShutdown(t *testing.T) {
	StartServer(multipleWorkersConfig())
	ShutdownServer()
}
