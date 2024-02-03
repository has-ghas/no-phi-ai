package scanner

import (
	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanOrganization struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a GitHub organization.
type ScanOrganization struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObject
}

// NewScanOrganization() function initializes a new ScanOrganization object using
// the provided URL for the GitHub organization.
func NewScanOrganization(
	org_url string,
	channel_documents chan<- az.AsyncDocumentWrapper,
	channel_quit <-chan error,
) (*ScanOrganization, error) {
	if scannerContext == nil {
		return nil, ErrScanOrganizationContextNil
	}

	// parse the name of the organization from the provided URL
	org_name, err := nogit.ParseOrgNameFromURL(org_url)
	if err != nil {
		return nil, err
	}

	// initialize and return a new ScanOrganization object
	return &ScanOrganization{
		ScanObject: *NewScanObject(&ScanObjectInput{
			ChannelDocuments: channel_documents,
			ChannelQuit:      channel_quit,
			ID:               org_url,
			Name:             org_name,
			ObjectType:       ScanObjectTypeOrganization,
			URL:              org_url,
		}),
	}, nil
}
