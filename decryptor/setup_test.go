/*
	Testing set up
*/

package decryptor

import (
	"github.com/mngharbi/DMPC/core"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log = core.InitializeLogging()
	log.SetLogLevel(core.WARN)
	shutdownProgram = core.ShutdownLambda(func() {})
	retCode := m.Run()
	os.Exit(retCode)
}
