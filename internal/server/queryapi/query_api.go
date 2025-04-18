package queryapi

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/armadaproject/armada/internal/common/compress"
	"github.com/armadaproject/armada/internal/common/database/lookout"
	protoutil "github.com/armadaproject/armada/internal/common/proto"
	"github.com/armadaproject/armada/internal/server/queryapi/database"
	"github.com/armadaproject/armada/pkg/api"
)

// JobStateMap is a mapping between database state and api Job states
var JobStateMap = map[int16]api.JobState{
	lookout.JobLeasedOrdinal:    api.JobState_LEASED,
	lookout.JobQueuedOrdinal:    api.JobState_QUEUED,
	lookout.JobPendingOrdinal:   api.JobState_PENDING,
	lookout.JobRunningOrdinal:   api.JobState_RUNNING,
	lookout.JobSucceededOrdinal: api.JobState_SUCCEEDED,
	lookout.JobFailedOrdinal:    api.JobState_FAILED,
	lookout.JobCancelledOrdinal: api.JobState_CANCELLED,
	lookout.JobPreemptedOrdinal: api.JobState_PREEMPTED,
	lookout.JobRejectedOrdinal:  api.JobState_REJECTED,
}

// JobRunStateMap is a mapping between database state and api Job Run states
var JobRunStateMap = map[int16]api.JobRunState{
	lookout.JobRunLeasedOrdinal:    api.JobRunState_RUN_STATE_LEASED,
	lookout.JobRunPendingOrdinal:   api.JobRunState_RUN_STATE_PENDING,
	lookout.JobRunRunningOrdinal:   api.JobRunState_RUN_STATE_RUNNING,
	lookout.JobRunSucceededOrdinal: api.JobRunState_RUN_STATE_SUCCEEDED,
	lookout.JobRunFailedOrdinal:    api.JobRunState_RUN_STATE_FAILED,
	// Lookout seems to have no concept of cancelling runs!
	// lookout.JobRunC:               api.JobRunState_RUN_STATE_CANCELLED,
	lookout.JobRunPreemptedOrdinal:     api.JobRunState_RUN_STATE_PREEMPTED,
	lookout.JobRunLeaseExpiredOrdinal:  api.JobRunState_RUN_STATE_LEASE_EXPIRED,
	lookout.JobRunLeaseReturnedOrdinal: api.JobRunState_RUNS_STATE_LEASE_RETURNED,
}

type QueryApi struct {
	db                  *pgxpool.Pool
	decompressorFactory func() compress.Decompressor
	maxQueryItems       int
}

func New(db *pgxpool.Pool, maxQueryItems int, decompressorFactory func() compress.Decompressor) *QueryApi {
	return &QueryApi{
		db:                  db,
		maxQueryItems:       maxQueryItems,
		decompressorFactory: decompressorFactory,
	}
}

func (q *QueryApi) GetJobDetails(ctx context.Context, req *api.JobDetailsRequest) (*api.JobDetailsResponse, error) {
	if len(req.JobIds) > q.maxQueryItems {
		return nil, fmt.Errorf("request contained more than %d jobIds", q.maxQueryItems)
	}

	detailsById := make(map[string]*api.JobDetails)

	err := pgx.BeginTxFunc(ctx, q.db, pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
		DeferrableMode: pgx.Deferrable,
	}, func(tx pgx.Tx) error {
		queries := database.New(tx)

		// Fetch the Job Rows
		resultRows, err := queries.GetJobDetails(ctx, req.JobIds)
		if err != nil {
			return err
		}

		jobsWithRuns := make([]string, 0, len(resultRows))
		decompressor := q.decompressorFactory()
		for _, row := range resultRows {
			var jobSpec *api.Job = nil
			if req.ExpandJobSpec {
				jobSpec, err = protoutil.DecompressAndUnmarshall[*api.Job](row.JobSpec, &api.Job{}, decompressor)
				if err != nil {
					return err
				}
			}
			apiJobState, ok := JobStateMap[row.State]
			if !ok {
				apiJobState = api.JobState_UNKNOWN
			}
			detailsById[row.JobID] = &api.JobDetails{
				JobId:            row.JobID,
				Queue:            row.Queue,
				Jobset:           row.Jobset,
				Namespace:        NilStringToString(row.Namespace),
				State:            apiJobState,
				SubmittedTs:      DbTimeToTimestamp(row.Submitted),
				CancelTs:         DbTimeToTimestamp(row.Cancelled),
				CancelReason:     NilStringToString(row.CancelReason),
				CancelUser:       NilStringToString(row.CancelUser),
				LastTransitionTs: DbTimeToTimestamp(row.LastTransitionTime),
				LatestRunId:      NilStringToString(row.LatestRunID),
				JobSpec:          jobSpec,
			}
			if req.GetExpandJobRun() && row.LatestRunID != nil {
				jobsWithRuns = append(jobsWithRuns, row.JobID)
			}
		}
		// Fetch the Job run details in a separate query.
		// We do this because each job can have many runs and so we don;t want to duplicate the job data for each run
		if len(jobsWithRuns) > 0 {
			runResultRows, err := queries.GetJobRunsByJobIds(ctx, jobsWithRuns)
			if err != nil {
				return err
			}
			runsByJob := make(map[string][]*api.JobRunDetails, len(resultRows))
			for _, row := range runResultRows {
				jobRuns, ok := runsByJob[row.JobID]
				if !ok {
					jobRuns = []*api.JobRunDetails{}
				}
				jobRuns = append(jobRuns, parseJobDetails(row))
				runsByJob[row.JobID] = jobRuns
			}

			for jobId, jobDetails := range detailsById {
				runs, ok := runsByJob[jobId]
				if ok {
					jobDetails.JobRuns = runs
				}
			}
		}
		return nil
	})

	return &api.JobDetailsResponse{
		JobDetails: detailsById,
	}, err
}

func (q *QueryApi) GetJobErrors(ctx context.Context, req *api.JobErrorsRequest) (*api.JobErrorsResponse, error) {
	if len(req.JobIds) > q.maxQueryItems {
		return nil, fmt.Errorf("request contained more than %d JobIds", q.maxQueryItems)
	}

	queries := database.New(q.db)
	queryResult, err := queries.GetJobErrorsByJobIds(ctx, req.JobIds)
	if err != nil {
		return nil, err
	}

	decompressor := q.decompressorFactory()
	errorsById := make(map[string]string, len(queryResult))
	for _, row := range queryResult {
		if len(row.Error) > 0 {
			decompressed, err := decompressor.Decompress(row.Error)
			if err != nil {
				return nil, err
			}
			errorsById[row.JobID] = string(decompressed)
		} else {
			errorsById[row.JobID] = ""
		}
	}
	return &api.JobErrorsResponse{
		JobErrors: errorsById,
	}, nil
}

func (q *QueryApi) GetJobRunDetails(ctx context.Context, req *api.JobRunDetailsRequest) (*api.JobRunDetailsResponse, error) {
	if len(req.RunIds) > q.maxQueryItems {
		return nil, fmt.Errorf("request contained more than %d RunIds", q.maxQueryItems)
	}

	queries := database.New(q.db)
	resultRows, err := queries.GetJobRunsByRunIds(ctx, req.RunIds)
	if err != nil {
		return nil, err
	}
	detailsById := make(map[string]*api.JobRunDetails, len(resultRows))
	for _, row := range resultRows {
		detailsById[row.RunID] = parseJobDetails(row)
	}
	return &api.JobRunDetailsResponse{
		JobRunDetails: detailsById,
	}, nil
}

func (q *QueryApi) GetJobStatus(ctx context.Context, req *api.JobStatusRequest) (*api.JobStatusResponse, error) {
	if len(req.JobIds) > q.maxQueryItems {
		return nil, fmt.Errorf("request contained more than %d jobIds", q.maxQueryItems)
	}

	queries := database.New(q.db)
	queryResult, err := queries.GetJobStates(ctx, req.JobIds)
	if err != nil {
		return nil, err
	}
	dbStatusById := make(map[string]int16, len(queryResult))
	for _, dbRow := range queryResult {
		dbStatusById[dbRow.JobID] = dbRow.State
	}

	apiStatusById := make(map[string]api.JobState, len(queryResult))
	for _, jobId := range req.JobIds {
		dbStatus, ok := dbStatusById[jobId]
		if ok {
			apiStatusById[jobId] = parseDbJobStateToApi(dbStatus)
		} else {
			apiStatusById[jobId] = api.JobState_UNKNOWN // We don't know about this job
		}
	}

	return &api.JobStatusResponse{
		JobStates: apiStatusById,
	}, nil
}

func (q *QueryApi) GetJobStatusUsingExternalJobUri(ctx context.Context, req *api.JobStatusUsingExternalJobUriRequest) (*api.JobStatusResponse, error) {
	if req.ExternalJobUri == "" {
		return nil, fmt.Errorf("request must contain external job uri")
	}

	queries := database.New(q.db)
	queryResult, err := queries.GetJobStatesUsingExternalSystemUri(ctx, database.GetJobStatesUsingExternalSystemUriParams{
		Queue:          req.Queue,
		Jobset:         req.Jobset,
		ExternalJobUri: req.ExternalJobUri,
	})
	if err != nil {
		return nil, err
	}

	apiStatusById := make(map[string]api.JobState, len(queryResult))
	for _, dbRow := range queryResult {
		apiStatusById[dbRow.JobID] = parseDbJobStateToApi(dbRow.State)
	}

	return &api.JobStatusResponse{
		JobStates: apiStatusById,
	}, nil
}

func (q *QueryApi) GetActiveQueues(ctx context.Context, _ *api.GetActiveQueuesRequest) (*api.GetActiveQueuesResponse, error) {
	queries := database.New(q.db)
	activeQueues, err := queries.GetActiveQueuesByPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get active queues by pool")
	}
	queuesByPool := map[string]*api.ActiveQueues{}
	for _, result := range activeQueues {
		if result.Pool == nil {
			continue
		}

		pool := *result.Pool
		if _, ok := queuesByPool[pool]; !ok {
			queuesByPool[pool] = &api.ActiveQueues{
				Queues: []string{},
			}
		}
		queuesByPool[pool].Queues = append(queuesByPool[pool].Queues, result.Queue)
	}

	return &api.GetActiveQueuesResponse{
		ActiveQueuesByPool: queuesByPool,
	}, nil
}

func parseDbJobStateToApi(dbStatus int16) api.JobState {
	apiStatus, ok := JobStateMap[dbStatus]
	if !ok {
		apiStatus = api.JobState_UNKNOWN // We know about this job but we can't map its state
	}
	return apiStatus
}

func parseJobDetails(row database.JobRun) *api.JobRunDetails {
	runState, ok := JobRunStateMap[row.JobRunState]
	if !ok {
		runState = api.JobRunState_RUN_STATE_UNKNOWN
	}
	return &api.JobRunDetails{
		RunId:      row.RunID,
		JobId:      row.JobID,
		State:      runState,
		Cluster:    row.Cluster,
		Node:       NilStringToString(row.Node),
		LeasedTs:   DbTimeToTimestamp(row.Leased),
		PendingTs:  DbTimeToTimestamp(row.Pending),
		StartedTs:  DbTimeToTimestamp(row.Started),
		FinishedTs: DbTimeToTimestamp(row.Finished),
	}
}
