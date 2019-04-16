package users

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Lambda to send a request and signers to users subsystem
*/
type Requester func(*core.VerifiedSigners, bool, bool, []byte) (chan *UserResponse, []error)

/*
	Logging
*/
var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Server API
*/

type Config struct {
	NumWorkers int
}

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func StartServer(
	conf Config,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) error {
	provisionServerOnce()
	if !serverSingleton.isInitialized {
		log = loggingHandler
		shutdownProgram = shutdownLambda
		serverSingleton.isInitialized = true
		serverHandler.ResetServer()
		serverHandler.InitServer(&serverSingleton)
	}
	return serverHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	provisionServerOnce()
	serverHandler.ShutdownServer()
}

func MakeUnverifiedRequest(signers *core.VerifiedSigners, readLock bool, readUnlock bool, rawRequest []byte) (chan *UserResponse, []error) {
	log.Debugf(receivedRequestLogMsg)
	return makeEncodedRequest(signers, readLock, readUnlock, rawRequest, true)
}

func MakeRequest(signers *core.VerifiedSigners, readLock bool, readUnlock bool, rawRequest []byte) (chan *UserResponse, []error) {
	log.Debugf(receivedRequestLogMsg)
	return makeEncodedRequest(signers, readLock, readUnlock, rawRequest, false)
}

// @TODO: lock/unlock needs to be moved out to locker subsystem
func makeEncodedRequest(signers *core.VerifiedSigners, readLock bool, readUnlock bool, rawRequest []byte, skipPermissions bool) (chan *UserResponse, []error) {
	// Build request object
	rqPtr := &UserRequest{}
	rqPtr.skipPermissions = skipPermissions
	decodingError := rqPtr.Decode(rawRequest)
	if decodingError != nil {
		return nil, []error{decodingError}
	}

	// Set issuer and certifier from arguments
	rqPtr.addSigners(signers)

	// Set read lock/unlock
	rqPtr.ReadLock = readLock
	rqPtr.ReadUnlock = readUnlock

	return makeRequest(rqPtr)
}

func makeRequest(rqPtr *UserRequest) (chan *UserResponse, []error) {
	// Sanitize request
	sanitizationErrors := rqPtr.sanitizeAndCheckParams()
	if len(sanitizationErrors) != 0 {
		return nil, sanitizationErrors
	}

	// Make request to server
	nativeResponseChannel, err := serverHandler.MakeRequest(rqPtr)
	if err != nil {
		return nil, []error{err}
	}

	// Pass through result
	responseChannel := make(chan *UserResponse)
	go func() {
		nativeResponse, ok := <-nativeResponseChannel
		if ok {
			responseChannel <- (*nativeResponse).(*UserResponse)
		} else {
			close(responseChannel)
		}
	}()

	return responseChannel, nil
}

/*
	Server implementation
*/

const (
	idIndexStr string = "id"
)

type server struct {
	isInitialized bool
	store         *memstore.Memstore
}

// Indexes used to store users
var indexesMap map[string]bool = map[string]bool{
	idIndexStr: true,
}

func getIndexes() (res []string) {
	for k := range indexesMap {
		res = append(res, k)
	}
	return res
}

var serverSingleton server
var serverHandler *gofarm.ServerHandler

func (sv *server) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		sv.store = memstore.New(getIndexes())
	}
	log.Debugf(daemonStartLogMsg)
	return nil
}

func (sv *server) Shutdown() error {
	log.Debugf(daemonShutdownLogMsg)
	return nil
}

func (sv *server) Work(request *gofarm.Request) *gofarm.Response {
	log.Debugf(runningRequestLogMsg)

	rq := (*request).(*UserRequest)

	/*
		Handle record level locking
	*/

	lockNeeds := []core.LockNeed{}

	// Add need for read locks for issuer and certifier
	if !rq.skipPermissions {
		lockNeeds = []core.LockNeed{
			{core.ReadLockType, rq.signers.IssuerId},
			{core.ReadLockType, rq.signers.CertifierId},
		}
	}

	// Add write lock for user record if updating
	if rq.Type == UpdateRequest {
		lockNeeds = append(lockNeeds, core.LockNeed{core.WriteLockType, rq.Data.Id})
	}

	// Add read locks for user records if reading
	lockNeedsWithoutSubjectReadLocks := lockNeeds
	if rq.Type == ReadRequest && rq.ReadLock {
		for _, userId := range rq.Fields {
			lockNeeds = append(lockNeeds, core.LockNeed{core.ReadLockType, userId})
		}
	}

	// Get locks needed (if locking)
	var userRecords []*userRecord
	isLocked := false
	if len(lockNeeds) != 0 {
		userRecords, isLocked = lockUsers(sv, lockNeeds)
	}

	// Read records if without read locking if option is on
	if rq.Type == ReadRequest && !rq.ReadLock {
		unlockedUserRecords := readUserRecordsByIds(sv.store, rq.Fields)
		for _, record := range unlockedUserRecords {
			userRecords = append(userRecords, record)
		}
	}

	// Find user records for certifier, issuer, and subject
	var issuerIndex, certifierIndex, subjectIndex int = -1, -1, -1
	for userRecordIndex, userRecord := range userRecords {
		if userRecord == nil {
			continue
		}

		if rq.signers != nil && userRecord.Id == rq.signers.IssuerId {
			issuerIndex = userRecordIndex
		}
		if rq.signers != nil && userRecord.Id == rq.signers.CertifierId {
			certifierIndex = userRecordIndex
		}
		if userRecord.Id == rq.Data.Id {
			subjectIndex = userRecordIndex
		}
	}

	// If any failed (not found), end job with corresponding failure
	if !isLocked {
		if !rq.skipPermissions && issuerIndex == -1 {
			return failRequest(IssuerUnknownError)
		}
		if !rq.skipPermissions && certifierIndex == -1 {
			return failRequest(CertifierUnknownError)
		}
		if subjectIndex == -1 && (rq.Type == ReadRequest || rq.Type == UpdateRequest) {
			return failRequest(SubjectUnknownError)
		}
	}

	/*
		Verify certifier permissions
	*/
	if !rq.skipPermissions {
		certifier := userRecords[certifierIndex]
		if !certifier.isAuthorized(rq) {
			// Unlock first
			_, isUnlocked := unlockUsers(sv, lockNeeds)
			if !isUnlocked {
				return failRequest(UnlockingFailedError)
			}
			// Then fail with certifier permissions error
			return failRequest(CertifierPermissionsError)
		}
	}

	/*
		Run request
	*/
	responseData := []*UserObject{}
	switch rq.Type {
	case UpdateRequest:
		// Determine memstore update mode
		isIndexUpdated := false
		for _, updatedFieldName := range rq.Fields {
			if indexesMap[updatedFieldName] {
				isIndexUpdated = true
				break
			}
		}

		// Make search record
		searchRecordPtr := (&rq.Data).makeSearchByIdRecord()

		// Atomically apply request to record in memstore
		updateFunc := func(obj memstore.Item) (memstore.Item, bool) {
			objCopy := obj.(*userRecord)
			objCopy.applyUpdateRequest(rq)
			return objCopy, true
		}
		var modifiedRecord *userRecord
		if isIndexUpdated {
			modifiedRecord = sv.store.UpdateWithIndexes(searchRecordPtr, "id", updateFunc).(*userRecord)
		} else {
			modifiedRecord = sv.store.UpdateData(searchRecordPtr, "id", updateFunc).(*userRecord)
		}

		// Add user modified to response
		modifiedObject := &UserObject{}
		modifiedObject.createFromRecord(modifiedRecord)
		responseData = append(responseData, modifiedObject)

	case CreateRequest:
		// Generate record
		newUser := &userRecord{
			lock: &sync.RWMutex{},
		}
		newUser.create(rq)

		// Add to memstore
		sv.store.Add(newUser)

		// Add user created to response
		createdObject := &UserObject{}
		createdObject.createFromRecord(newUser)
		responseData = append(responseData, createdObject)

	case ReadRequest:
		// Extract indexes for users requested
		usersRequestedIds := []int{}
		for _, userId := range rq.Fields {
			for userRecordIndex, userRecord := range userRecords {
				if userRecord.Id == userId {
					usersRequestedIds = append(usersRequestedIds, userRecordIndex)
					break
				}
			}
		}

		// Transform records requested into objects and add to response
		for _, userRecordIndex := range usersRequestedIds {
			var createdObject *UserObject = nil
			if userRecords[userRecordIndex] != nil {
				createdObject = &UserObject{}
				createdObject.createFromRecord(userRecords[userRecordIndex])
			}
			responseData = append(responseData, createdObject)
		}
	}

	/*
		Handle unlocking
	*/

	// Add lock needs if read unlocking without locking
	if rq.Type == ReadRequest && !rq.ReadLock && rq.ReadUnlock {
		for _, userId := range rq.Fields {
			lockNeeds = append(lockNeeds, core.LockNeed{core.ReadLockType, userId})
		}
	}

	// Use lock needs without subject locks if locking without unlocking
	if rq.Type == ReadRequest && rq.ReadLock && !rq.ReadUnlock {
		lockNeeds = lockNeedsWithoutSubjectReadLocks
	}

	// Do unlocking
	if len(lockNeeds) != 0 {
		if _, isUnlocked := unlockUsers(sv, lockNeeds); !isUnlocked {
			return failRequest(UnlockingFailedError)
		}
	}

	// Request is done, return response generated
	return successRequest(responseData)
}

func failRequest(responseCode int) *gofarm.Response {
	log.Debugf(failRequestLogMsg)
	userRespPtr := &UserResponse{
		Result: responseCode,
		Data:   []UserObject{},
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}

func successRequest(responseData []*UserObject) *gofarm.Response {
	log.Debugf(successRequestLogMsg)
	var objectDataCopy []UserObject
	for _, objectPtr := range responseData {
		objectDataCopy = append(objectDataCopy, *objectPtr)
	}

	userRespPtr := &UserResponse{
		Result: Success,
		Data:   objectDataCopy,
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}
