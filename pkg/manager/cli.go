package manager

import (
	"github.com/pkg/errors"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/scanner"
	"github.com/has-ghas/no-phi-ai/pkg/scannerv2"
)

// initCLI() method is used to initialize clients used by the Manager in the
// running of CLI commands
func (m *Manager) initCLI() (e error) {
	switch m.config.Command.Run {
	case cfg.CommandRunHelp:
		// no clients to initialize for the help command
		return
	case cfg.CommandRunScanTest:
		m.scanner_v2, e = scannerv2.NewScanner(m.ctx, m.config, scannerv2.NewMemoryResultRecordIO(m.ctx))
		if e != nil {
			e = errors.Wrapf(e, "failed to initialize new Scanner for command %s", m.config.Command.Run)
			return
		}
		return
	case cfg.CommandRunVersion:
		// no clients to initialize for the version command
		return
	default:
		// initialize the Scanner for use by the Manager in the running of CLI
		// commands that involve scanning git repos.
		m.scanner, e = scanner.NewScanner(m.config, m.ctx, m.logger)
		if e != nil {
			e = errors.Wrapf(e, "failed to initialize new Scanner for command %s", m.config.Command.Run)
			return
		}
		return
	}
}

// runCLI() method is used to run the command specified in m.config.Command.Run var.
func (m *Manager) runCLI() (e error) {
	switch m.config.Command.Run {
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
	case cfg.CommandRunScanTest:
		e = m.commandScanTest()
		return
	case cfg.CommandRunVersion:
		e = m.commandVersion()
		return
	default:
		m.logger.Fatal().Msgf("invalid command : %s", m.config.Command.Run)
		return
	}
}
