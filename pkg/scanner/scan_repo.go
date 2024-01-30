package scanner

import (
	git "github.com/go-git/go-git/v5"
)

// ScanRepository struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a git.Repository.
type ScanRepository struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObject

	repository *git.Repository
}

// NewScanRepository() function initializes a new ScanRepository object.
func NewScanRepository(name, url string) *ScanRepository {
	return &ScanRepository{
		ScanObject: *NewScanObject(name, "repository", url),
	}
}

// GetRepository() method returns a pointer to the git.Repository
// associated with the ScanRepository.
func (sr *ScanRepository) GetRepository() *git.Repository {
	return sr.repository
}

// SetRepository() method stores a pointer to the git.Repository
// associated with the ScanRepository.
func (sr *ScanRepository) SetRepository(repo *git.Repository) {
	sr.repository = repo
}
