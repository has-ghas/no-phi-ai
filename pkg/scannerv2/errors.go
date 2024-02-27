package scannerv2

import "github.com/pkg/errors"

const (
	ErrMsgAddScanRepository        = "failed to add ScanRepository"
	ErrMsgResultWriteFailed        = "failed to write result"
	ErrMsgScanFileReader           = "failed to create reader for file"
	ErrMsgScanFileRequestsGenerate = "failed to generate new requests for file %s"
	ErrMsgScanRepositoryCreate     = "failed to create new ScanRepository"
	ErrMsgScanTrackerUpdateFile    = "failed to update tracker for file %s"
	ErrMsgScannerCreate            = "failed to create new Scanner"
)

var (
	ErrChunkFileToRequestsFailed         = errors.New("failed to chunk non-empty file into one or more requests")
	ErrDocumentResponseMismatch          = errors.New("DocumentResponse mismatch")
	ErrKeyAddKeyEmpty                    = errors.New("cannot add key : key is empty")
	ErrKeyAddKeyExists                   = errors.New("cannot add key : key already exists")
	ErrKeyCodeInvalid                    = errors.New("invalid key code")
	ErrKeyTrackerInvalidKind             = errors.New("invalid kind for KeyTracker")
	ErrKeyUpdateKeyEmpty                 = errors.New("cannot update key : key is empty")
	ErrMemoryResultRecordIODeleteEmptyID = errors.New("memory store failed to delete result record : empty ID")
	ErrMemoryResultRecordIOReadEmptyID   = errors.New("memory store failed to read result record : empty ID")
	ErrMemoryResultRecordIOReadFailed    = errors.New("memory store failed to read result record")
	ErrNewRequestEmptyCommitID           = errors.New("cannot create a new request with an empty commit ID")
	ErrNewRequestEmptyObjectID           = errors.New("cannot create a new request with an empty object ID")
	ErrNewRequestEmptyRepositoryID       = errors.New("cannot create a new request with an empty repository ID")
	ErrNewRequestEmptyText               = errors.New("cannot create a new request with an empty text")
	ErrProcessRequestNoID                = errors.New("cannot process a request without a valid ID")
	ErrProcessResponseNoID               = errors.New("cannot process a response without a valid ID")
	ErrScannerAddScanRepositoryEmptyID   = errors.New("cannot add a ScanRepository with an empty ID")
	ErrScannerAddScanRepositoryNil       = errors.New("cannot add a nil ScanRepository to scanner")
	ErrScannerGetScanRepositoryNotFound  = errors.New("ScanRepository not found")
	ErrScanRepositoryChannelDocumentsNil = errors.New("ScanRepository channelDocuments channel is nil")
	ErrScanRepositoryConfigNil           = errors.New("ScanRepository config is nil")
	ErrScanRepositoryContextNil          = errors.New("ScanRepository requires a non-nil context")
	ErrScanRepositoryCloneGitManagerNil  = errors.New("ScanRepository git manager is nil")
	ErrScanRepositorySetRepositoryNil    = errors.New("ScanRepository cannot set nil repository")
)
