package locker

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
)

/*
	Error messages
*/

const (
	unknownResourceTypeErrorMsg string = "Unknown resource type"
	noNeedsErrorMsg             string = "Request needs to request at least one resource"
)

/*
	Definition of resource type
*/

type ResourceType int

const ChannelLock ResourceType = 0

/*
	Structure of a locker request
*/

type LockerRequest struct {
	Type        ResourceType
	Needs       []core.LockNeed
	LockingType core.LockingType
}

// Used to check errors within request
func (rq *LockerRequest) checkAndPrepareRequest() []error {
	res := []error{}

	// Check type
	if rq.Type != ChannelLock {
		res = append(res, errors.New(unknownResourceTypeErrorMsg))
	}

	// Check we have at least one request
	if len(rq.Needs) == 0 {
		res = append(res, errors.New(noNeedsErrorMsg))
	}

	return res
}
