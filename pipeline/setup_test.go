/*
	Testing set up
*/

package pipeline

import (
	"github.com/mngharbi/DMPC/core"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log = core.InitializeLogging()
	log.SetLogLevel(core.WARN)
	retCode := m.Run()
	os.Exit(retCode)
}
