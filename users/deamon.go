package users

import (
	"github.com/mngharbi/memstore"
	"github.com/mngharbi/gofarm"
)


/*
	Server API
*/

type Config struct {
	NumWorkers		int
}

func StartServer(conf Config) {
	if !serverSingleton.isInitialized {
		serverSingleton.isInitialized = true
		gofarm.InitServer(&serverSingleton)
	}
	gofarm.StartServer(gofarm.Config{ NumWorkers: conf.NumWorkers })
}

func ShutdownServer() {
	gofarm.ShutdownServer()
}

func MakeRequest(rawRequest []byte) (chan *UserResponse, []error) {
	// Build request object
	rqPtr := &UserRequest{}
	decodingErrors := rqPtr.Decode(rawRequest)
	if len(decodingErrors) > 0 {
		return nil, decodingErrors
	}

	// Make request to server and pass through result
	nativeResponseChannel, _ := gofarm.MakeRequest(rqPtr)
	var responseChannel chan *UserResponse
	go func() {
		nativeResponse, ok := <- nativeResponseChannel
		if ok {
			responseChannel <- (*nativeResponse).(*UserResponse)
		} else {
			close(responseChannel)
		}
	}()

	return responseChannel, []error{}
}

/*
	Server implementation
*/

type server struct {
	isInitialized	bool
	store			*memstore.Memstore
}

// Indexes used to store users
var indexesMap map[string]bool = map[string]bool{
	"id": true,
}

func getIndexes() (res []string) {
	for k := range indexesMap {
		res = append(res, k)
	}
	return res
}

var serverSingleton server

func (sv *server) Start (_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if !isFirstStart {
		sv.store = memstore.New(getIndexes())
	}
	return nil
}

func (sv *server) Shutdown() error { return nil }

func (sv *server) Work (request *gofarm.Request) *gofarm.Response {
	rq := (*request).(*UserRequest)

	/*
		Handle record level locking
	*/

	// Add need for read locks for issuer and certifier
	lockNeeds := []lockNeed{
		lockNeed{false, rq.IssuerId},
		lockNeed{false, rq.CertifierId},
	}

	// Add write lock for user record if updating
	if rq.Type == UpdateRequest {
		lockNeeds = append(lockNeeds, lockNeed{true, rq.Data.Id})
	}

	// Add read locks for user records if reading
	if rq.Type == ReadRequest {
		for _,userId := range rq.Fields {
			lockNeeds = append(lockNeeds, lockNeed{false, userId})
		}
	}

	// Get locks needed
	userRecords, isLocked := lockUsers(sv, lockNeeds)

	// Find user records for certifier, issuer, and subject
	var issuerIndex, certifierIndex, subjectIndex int = -1, -1, -1
	for userRecordIndex, userRecord := range userRecords {
		if userRecord == nil {
			continue
		}

		switch userRecord.Id {
		case rq.IssuerId:
			issuerIndex = userRecordIndex
		case rq.CertifierId:
			certifierIndex = userRecordIndex
		case rq.Data.Id:
			subjectIndex = userRecordIndex
		}
	}

	// If any failed (not found), end job with corresponding failure
	if !isLocked {
		if issuerIndex == -1 {
			return failRequest(IssuerUnknownError)
		}
		if certifierIndex == -1 {
			return failRequest(CertifierUnknownError)
		}
		if subjectIndex == -1 && rq.Type == ReadRequest {
			return failRequest(SubjectUnknownError)
		}
	}

	/*
		Verify certifier permisisons
	*/
	certifier := userRecords[certifierIndex]
	if !certifier.isAuthorized(rq) {
		return failRequest(CertifierPermissionsError)
	}

	/*
		Run request
	*/
	responseData := []*UserObject{}
	switch(rq.Type) {
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
		searchRecordPtr := makeSearchByIdRecord(&rq.Data)

		// Atomically apply request to record in memstore
		updateFunc := func(obj memstore.Item) (memstore.Item, bool) {
			objCopy := obj.(userRecord)
			objCopy.applyUpdateRequest(rq)
			return objCopy, true
		}
		var modifiedRecord *userRecord
		if isIndexUpdated {
			*modifiedRecord = sv.store.UpdateWithIndexes(*searchRecordPtr, "id", updateFunc).(userRecord)
		} else {
			*modifiedRecord = sv.store.UpdateData(*searchRecordPtr, "id", updateFunc).(userRecord)
		}

		// Add user modified to response
		modifiedObject := &UserObject{}
		modifiedObject.createFromRecord(modifiedRecord)
		responseData = append(responseData, modifiedObject)


	case CreateRequest:
		// Generate record
		newUser := &userRecord{}
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
		for _,userId := range rq.Fields {
			for userRecordIndex, userRecord := range userRecords {
				if userRecord.Id == userId {
					usersRequestedIds = append(usersRequestedIds, userRecordIndex)
					break
				}
			}
		}

		// Transform records requested into objects and add to response
		for _, userRecordIndex := range usersRequestedIds {
			createdObject := &UserObject{}
			createdObject.createFromRecord(userRecords[userRecordIndex])
			responseData = append(responseData, createdObject)
		}
	}

	/*
		Handle unlocking
	*/
	unlockUsers(sv, lockNeeds)

	// Request is done, return response generated
	return successRequest(responseData)
}

func failRequest(responseCode int) *gofarm.Response {
	userRespPtr := &UserResponse{
		Result: responseCode,
		Data: []UserObject{},
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}

func successRequest(responseData []*UserObject) *gofarm.Response {
	var objectDataCopy []UserObject
	for _,objectPtr := range responseData {
		objectDataCopy = append(objectDataCopy, *objectPtr)
	}

	userRespPtr := &UserResponse{
		Result: Success,
		Data: objectDataCopy,
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}