package users

import (
	"github.com/mngharbi/memstore"
	"sync"
)


/*
	Server API
*/

type Config struct {
	NumWorkers		int
}

func StartServer(conf Config) {
	sv.init(&conf)
}

func ShutdownServer() {
	sv.shutdown()
}

func MakeRequest(request []byte) (chan UserResponse) {
	// Build corresponding job and push it into server's incoming job pipe
	var reqJob job
	reqJob.response = make(chan *UserResponse)
	reqJob.request = make([]byte, len(request))
	copy(reqJob.request, request)
	sv.jobPipe <- &reqJob

	// Pass response through
	var publicResponse chan UserResponse
	go func() {
		var response *UserResponse
		response = <- reqJob.response
		publicResponse <- *response
	}()

	return publicResponse
}

/*
	Server implementation
*/

type server struct {
	isInitialized	bool
	workerPool		[]*worker
	freeWorkers		[]*worker
	jobPool			[]*job
	jobPipe			chan *job
	workerPipe		chan *worker
	shutdownPipe	chan bool
	workerWaitGroup	sync.WaitGroup
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

var sv server

func (sv *server) init (conf *Config) {
	// Setup channels
	sv.jobPipe = make(chan *job)
	sv.workerPipe = make(chan *worker)
	sv.shutdownPipe = make(chan bool)

	// Build workers
	sv.workerPool = make([]*worker, conf.NumWorkers)
	sv.freeWorkers = make([]*worker, conf.NumWorkers)
	for workerIndex := 0; workerIndex < conf.NumWorkers ; workerIndex++ {
		newWorker := worker{
			jobs: make(chan *job),
			shutdownPipe: make(chan bool),
		}
		sv.workerPool[workerIndex] = &newWorker
		sv.freeWorkers[workerIndex] = &newWorker
	}

	// Startup workers
	sv.workerWaitGroup.Add(conf.NumWorkers)
	for _,worker := range sv.workerPool {
		go worker.run()
	}

	// Initialize store (only if starting for the first time)
	if !sv.isInitialized {
		sv.store = memstore.New(getIndexes())
	}

	// Start running server
	go sv.run()

	// Set flag for restarts
	sv.isInitialized = true
}

func (sv *server) run () {
	running := true
	for running {
		select {
		// Accept incoming jobs
		case newJob := <- sv.jobPipe:
			sv.jobPool = append(sv.jobPool, newJob)

		// Accept status update from worker
		case freeWorker := <- sv.workerPipe:
			sv.freeWorkers = append(sv.freeWorkers, freeWorker)

		// Shutdown
		case <- sv.shutdownPipe:
			// Signal workers to shutdown and wait for them
			for _,worker := range sv.workerPool {
				go worker.shutdown()
			}
			sv.workerWaitGroup.Wait()

			// Reset defaults
			sv.workerPool = []*worker{}
			sv.freeWorkers = []*worker{}
			sv.jobPool = []*job{}

			running = false
			break
		}

		// Distribute available jobs to avaiable workers in order
		sv.distributeJobs()
	}
}

func (sv *server) distributeJobs () {
	// Count maximum amount of assignments possible
	jobPoolLen := len(sv.jobPool)
	freeWorkersLen := len(sv.freeWorkers)
	assignmentsCount := jobPoolLen
	if freeWorkersLen < jobPoolLen {
		assignmentsCount = freeWorkersLen
	}
	if assignmentsCount == 0 {
		return
	}

	// Remove jobs/workers to be assigned
	jobsToAssign := sv.jobPool[:assignmentsCount]
	workersToAssign := sv.freeWorkers[:assignmentsCount]
	sv.jobPool = sv.jobPool[assignmentsCount:]
	sv.freeWorkers = sv.freeWorkers[assignmentsCount:]

	// Put jobs into their worker's incoming channel
	for i := 0; i < assignmentsCount; i++ {
		workersToAssign[i].jobs <- jobsToAssign[i]
	}
}

func (sv *server) shutdown () {
	sv.shutdownPipe <- true
}

/*
	Worker implementation
*/

type worker struct {
	jobs			chan *job
	shutdownPipe	chan bool
}

type job struct {
	response	chan *UserResponse
	request		[]byte
}

func (wk *worker) run () {
	running := true
	for running {
		select {
		// Accept incoming jobs
		case newJob := <- wk.jobs:
			// Do job
			wk.doJob(newJob)

			// Report to server that we're free
			sv.workerPipe <- wk

		// Shutdown
		case <- wk.shutdownPipe:
			running = false
			break
		}
	}

	sv.workerWaitGroup.Done()
}



func (wk *worker) doJob (job *job) {
	/*
		Decoding, request sanitization
	*/
	rq := &UserRequest{}
	decodingErrors := rq.Decode(job.request)

	if len(decodingErrors) > 0 {
		wk.failJob(job, DecodeError)
		return
	}

	/*
		Handle record level locking
	*/

	// Add need for read locks for issuer and certifier
	lockNeeds := []lockNeed{
		lockNeed{false, rq.IssuerId},
		lockNeed{false, rq.CertifierId},
	}

	// Determine whether we need write lock for user object
	if rq.Type == UpdateRequest {
		lockNeeds = append(lockNeeds, lockNeed{true, rq.Data.Id})
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
			wk.failJob(job, IssuerUnknownError)
			return
		}
		if certifierIndex == -1 {
			wk.failJob(job, CertifierUnknownError)
			return
		}
		if subjectIndex == -1 {
			wk.failJob(job, SubjectUnknownError)
			return
		}
	}

	/*
		Verify certifier permisisons
	*/
	certifier := userRecords[certifierIndex]
	if !certifier.isAuthorized(rq) {
		wk.failJob(job, CertifierPermissionsError)
		return
	}


	/*
		Run request
	*/
	responseData := []*UserObject{}
	switch(rq.Type) {
	case UpdateRequest:
		// Determine memstore update mode
		isIndexUpdated := false
		for _, updatedFieldName := range rq.FieldsUpdated {
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
	}

	/*
		Handle unlocking
	*/
	unlockUsers(sv, lockNeeds)

	// Request is done, return response generated
	wk.successJob(job, responseData)
}

func (wr *worker) failJob(job *job, responseCode int) {
	job.response <- &UserResponse{
		Result: responseCode,
		Data: []UserObject{},
	}
}

func (wr *worker) successJob(job *job, responseData []*UserObject) {
	var objectDataCopy []UserObject
	for _,objectPtr := range responseData {
		objectDataCopy = append(objectDataCopy, *objectPtr)
	}

	job.response <- &UserResponse{
		Result: Success,
		Data: objectDataCopy,
	}
}

func (wk *worker) shutdown () {
	wk.shutdownPipe <- true
}
