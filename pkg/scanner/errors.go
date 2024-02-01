package scanner

import "github.com/pkg/errors"

var (
	ErrDocumentResponseIsNil            = errors.New("DocumentResponse is nil")
	ErrDocumentResponseMismatch         = errors.New("DocumentResponse mismatch")
	ErrScanCommitFilesNotSet            = errors.New("ScanCommit files not set")
	ErrScanCommitScanFileNotFound       = errors.New("ScanFile not found in ScanCommit")
	ErrScanFileInputNil                 = errors.New("ScanFile input is nil")
	ErrScanObjectDocumentsNotSet        = errors.New("ScanObject documents not set")
	ErrScanObjectDocumentsNotValid      = errors.New("ScanObject documents not valid")
	ErrScanRepositoryCommitsNotSet      = errors.New("ScanRepository commits not set")
	ErrScanRepositoryScanCommitNotFound = errors.New("ScanCommit not found in ScanRepository")
)
