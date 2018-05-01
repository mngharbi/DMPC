/*
	Test helpers
*/

package executor

import (
	"github.com/mngharbi/DMPC/core"
	"testing"
)

/*
	General
*/

const (
	genericIssuerId    string = "ISSUER_ID"
	genericCertifierId string = "CERTIFIER_ID"
)

func generateSigners(issuerId string, certifierId string) *core.VerifiedSigners {
	return &core.VerifiedSigners{
		IssuerId:    issuerId,
		CertifierId: certifierId,
	}
}

func generateGenericSigners() *core.VerifiedSigners {
	return generateSigners(genericIssuerId, genericCertifierId)
}

/*
	Server
*/

func resetAndStartServer(
	t *testing.T,
	conf Config,
	usersRequester UsersRequester,
	usersRequesterUnverified UsersRequester,
	responseReporter ResponseReporter,
	ticketGenerator TicketGenerator,
) bool {
	serverSingleton = server{}
	InitializeServer(usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator, log, shutdownProgram)
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
	}
}
