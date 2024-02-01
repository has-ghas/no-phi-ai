package scanner

import nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"

// ScanOrganization struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a GitHub organization.
type ScanOrganization struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObject
}

// NewScanOrganization() function initializes a new ScanOrganization object using
// the provided URL for the GitHub organization.
func NewScanOrganization(org_url string) (*ScanOrganization, error) {
	// parse the name of the organization from the provided URL
	org_name, err := nogit.ParseOrgNameFromURL(org_url)
	if err != nil {
		return nil, err
	}
	// initialize and return a new ScanOrganization object
	return &ScanOrganization{
		ScanObject: *NewScanObject(
			org_url,
			org_name,
			ScanObjectTypeOrganization,
			org_url,
		),
	}, nil
}