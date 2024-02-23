package manager

import (
	"fmt"

	"github.com/pkg/errors"

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
		"... work in progress ...",
	)
	printNameAndDescription(
		cfg.CommandRunScanTest,
		"... work in progress ...",
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

// commandScanTest() method is used to run the "scan-test" command, which is
// for development use only.
func (m *Manager) commandScanTest() (e error) {

	if len(m.config.Git.Scan.Repositories) == 0 {
		e = errors.New("no repositories specified for scan")
		return
	}

	err_chan := make(chan error)
	go m.scanner_v2.Run(err_chan)

	// wait for an error to be returned from the scanner
	e = <-err_chan
	if e != nil {
		e = errors.Wrapf(e, "failed to run command '%s' ", m.config.Command.Run)
		return
	}
	m.logger.Info().Msgf("command '%s' completed successfully", m.config.Command.Run)

	return
}

// commandVersion() method is used to run the "version" command, which prints
// the version information for the app and then exits.
func (m *Manager) commandVersion() (e error) {
	fmt.Printf("%s %s\n", m.config.App.Name, cfg.AppVersion)
	return
}

// printNameAndDescription() helper function is used to print the name and (optional)
// description of something, such as a command or environment variable, to stdout.
func printNameAndDescription(name string, description string) {
	if description != "" {
		fmt.Printf("\t\t- %s : %s\n", name, description)
	} else {
		fmt.Printf("\t\t- %s\n", name)
	}
}
