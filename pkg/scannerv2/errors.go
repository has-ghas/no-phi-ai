package scannerv2

import "github.com/pkg/errors"

const (
	ErrMsgAddScanRepository     = "failed to add ScanRepository"
	ErrMsgResultWriteFailed     = "failed to write result"
	ErrMsgScanRepositoryCreate  = "failed to create new ScanRepository"
	ErrMsgScanTrackerUpdateFile = "failed to update tracker for file %s"
	ErrMsgScannerCreate         = "failed to create new Scanner"
	ErrMsgTrackerUpdateCommit   = "failed to update tracker for commit %s"
)

var (
	ErrKeyAddKeyEmpty                    = errors.New("cannot add key : key is empty")
	ErrKeyAddKeyExists                   = errors.New("cannot add key : key already exists")
	ErrKeyCodeInvalid                    = errors.New("invalid key code")
	ErrKeyTrackerInvalidKind             = errors.New("invalid kind for KeyTracker")
	ErrKeyUpdateKeyEmpty                 = errors.New("cannot update key : key is empty")
	ErrMemoryResultRecordIODeleteEmptyID = errors.New("memory store failed to delete result record : empty ID")
	ErrMemoryResultRecordIOReadEmptyID   = errors.New("memory store failed to read result record : empty ID")
	ErrMemoryResultRecordIOReadFailed    = errors.New("memory store failed to read result record")
	ErrProcessRequestNoID                = errors.New("cannot process a request without a valid ID")
	ErrProcessResponseNoID               = errors.New("cannot process a response without a valid ID")
	ErrScannerAddScanRepositoryEmptyID   = errors.New("cannot add a ScanRepository with an empty ID")
	ErrScannerAddScanRepositoryNil       = errors.New("cannot add a nil ScanRepository to scanner")
	ErrScannerGetScanRepositoryNotFound  = errors.New("ScanRepository not found")
	ErrScanRepositoryChannelErrorsNil    = errors.New("ScanRepository errors channel is nil")
	ErrScanRepositoryChannelRequestsNil  = errors.New("ScanRepository requests channel is nil")
	ErrScanRepositoryConfigNil           = errors.New("ScanRepository config is nil")
	ErrScanRepositoryContextNil          = errors.New("ScanRepository requires a non-nil context")
	ErrScanRepositoryCloneGitManagerNil  = errors.New("ScanRepository git manager is nil")
	ErrScanRepositorySetRepositoryNil    = errors.New("ScanRepository cannot set nil repository")
)
