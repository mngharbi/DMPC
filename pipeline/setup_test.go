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
	log.SetLogLevel(core.DEBUG)
	retCode := m.Run()
	os.Exit(retCode)
}
