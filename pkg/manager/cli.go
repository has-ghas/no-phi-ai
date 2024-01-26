package manager

import (
	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

func (m *Manager) initCLI() {
	// initialize the GitManager for use in the execution of CLI commands
	m.Git = nogit.NewGitManager(m.Config, m.Logger)
}

// runCLI() method is used to run the command specified in m.Config.Command.Run var.
func (m *Manager) runCLI() (e error) {
	switch m.Config.Command.Run {
	case cfg.CommandRunHelp:
		e = m.commandHelp()
		return
	case cfg.CommandRunListOrgRepos:
		e = m.commandListOrgRepos()
		return
	case cfg.CommandRunScanOrg:
		e = m.commandScanOrg()
		return
	case cfg.CommandRunScanRepos:
		e = m.commandScanRepos()
		return
	case cfg.CommandRunVersion:
		e = m.commandVersion()
		return
	default:
		m.Logger.Fatal().Msgf("invalid command : %s", m.Config.Command.Run)
		return
	}
}
