package scanner

import "github.com/pkg/errors"

var (
	ErrDocumentResponseIsNil             = errors.New("DocumentResponse is nil")
	ErrDocumentResponseMismatch          = errors.New("DocumentResponse mismatch")
	ErrDocumentTrackerContextNil         = errors.New("DocumentTracker requires a non-nil context")
	ErrDocumentTrackerMapInputIsNil      = errors.New("DocumentTracker map input is nil")
	ErrScanCommitChannelDocumentsNil     = errors.New("ScanCommit channelDocuments channel is nil")
	ErrScanCommitInputCommitNil          = errors.New("ScanCommit in input Commit pointer is nil")
	ErrScanCommitContextNil              = errors.New("ScanCommit requires a non-nil context")
	ErrScanCommitFilesNotSet             = errors.New("ScanCommit files not set")
	ErrScanCommitScanFileNotFound        = errors.New("ScanFile not found in ScanCommit")
	ErrScanFileChannelDocumentsNil       = errors.New("ScanFile channelDocuments channel is nil")
	ErrScanFileContextNil                = errors.New("ScanFile requires a non-nil context")
	ErrScanFileInputNil                  = errors.New("ScanFile input is nil")
	ErrScanObjectDocumentsNotSet         = errors.New("ScanObject documents not set")
	ErrScanObjectDocumentsNotValid       = errors.New("ScanObject documents not valid")
	ErrScanOrganizationContextNil        = errors.New("ScanOrganization requires a non-nil context")
	ErrScanRepositoryChannelDocumentsNil = errors.New("ScanRepository channelDocuments channel is nil")
	ErrScanRepositoryContextNil          = errors.New("ScanRepository requires a non-nil context")
	ErrScanRepositoryCommitsNotSet       = errors.New("ScanRepository commits not set")
	ErrScanRepositoryScanCommitNotFound  = errors.New("ScanCommit not found in ScanRepository")
	ErrScanTrackerChannelDocumentsNil    = errors.New("ScanTracker channelDocuments channel is nil")
	ErrScanTrackerContextNil             = errors.New("ScanTracker requires a non-nil context")
	ErrScanTrackerGitManagerNil          = errors.New("ScanTracker gitManager cannot be nil")
)
