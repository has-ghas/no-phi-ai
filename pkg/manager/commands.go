package manager

import (
	"errors"
	"fmt"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
)

func (m *Manager) commandHelp() (e error) {
	fmt.Printf("CLI Help Information for %s app:\n", m.config.App.Name)
	fmt.Println("\tAvailable Commands:")
	printNameAndDescription(
		cfg.CommandRunHelp,
		"Prints (this) help information for the app.",
	)
	printNameAndDescription(
		cfg.CommandRunListOrgRepos,
		"... not yet implemented ...",
	)
	printNameAndDescription(
		cfg.CommandRunScanOrg,
		"... not yet implemented ...",
	)
	printNameAndDescription(
		cfg.CommandRunScanRepos,
		"... not yet implemented ...",
	)
	printNameAndDescription(
		cfg.CommandRunVersion,
		"Prints version information for the app.",
	)
	fmt.Println("\tEnvironment Variables:")
	for _, envVar := range cfg.GetAppEnvVars() {
		printNameAndDescription(envVar, "")
	}
	return
}

// commandListOrgRepos() method is used to run the "list-org-repos" command.
func (m *Manager) commandListOrgRepos() (e error) {
	m.logger.Warn().Msgf("%s commmand is TODO\n", cfg.CommandRunListOrgRepos)
	return
}

// commandScanOrg() method is used to run the "scan-org" command, which
// is applies the "scan-repos" command to all repositories in the organization.
func (m *Manager) commandScanOrg() (e error) {
	m.logger.Warn().Msgf("%s commmand is TODO\n", cfg.CommandRunScanOrg)
	return
}

// commandScanRepos() method is used to run the "scan-repos" command, which
// is used to scan the contents of a single git repository for PHI/PII.
func (m *Manager) commandScanRepos() (e error) {

	if len(m.config.Git.Scan.Repositories) == 0 {
		e = errors.New("no repositories specified for scan")
		return
	}

	e = m.scanner.ScanReposForPHI()
	return
}

func (m *Manager) commandVersion() (e error) {
	fmt.Printf("%s %s\n", m.config.App.Name, cfg.AppVersion)
	return
}

func printNameAndDescription(name string, description string) {
	if description != "" {
		fmt.Printf("\t\t- %s : %s\n", name, description)
	} else {
		fmt.Printf("\t\t- %s\n", name)
	}
}
