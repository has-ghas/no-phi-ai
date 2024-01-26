package nogit

import (
	"errors"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/rs/zerolog"
)

// GitManager struct provides a management wrapper for interactin
// with raw git repositories using the go-git library. Provides
// methods for cloning a repository and scanning for PHI/PII by
// recursively walking the repository's file tree.
type GitManager struct {
	Config *cfg.Config
	Logger *zerolog.Logger
}

// NewGitManager returns a new GitManager instance.
func NewGitManager(config *cfg.Config, logger *zerolog.Logger) *GitManager {
	return &GitManager{
		Config: config,
		Logger: logger,
	}
}

func (gm *GitManager) CloneRepo(repo_url string) (*git.Repository, error) {

	var key_err error
	var auth_method transport.AuthMethod
	auth_method, key_err = gm.getAuthMethod(repo_url)
	if key_err != nil {
		return nil, key_err
	}

	clone_dir, dir_err := gm.getRepoCloneDir(repo_url)
	if dir_err != nil {
		return nil, dir_err
	}

	gm.Logger.Debug().Msgf("cloning git repo %s to %s", repo_url, clone_dir)
	repo, err := git.PlainClone(clone_dir, false, &git.CloneOptions{
		Progress: os.Stdout,
		URL:      repo_url,
		Auth:     auth_method,
	})

	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			gm.Logger.Info().Msg("repo was already cloned")
			return git.PlainOpen(clone_dir)
		} else {
			gm.Logger.Error().Err(err).Msg("clone git repo error")
			return nil, err
		}
	}
	gm.Logger.Info().Msgf("cloned git repo to %s", clone_dir)

	return repo, nil
}

func (gm *GitManager) getAuthMethod(repo_url string) (transport.AuthMethod, error) {
	// use the provided config values to determine which auth method to use
	//
	// TODO : also use the repo_url to determine which auth method to use
	if gm.Config.GitHub.Auth.SSHKeyPath != "" {
		// use SSH key auth if configured
		return gm.getAuthMethodPublicKey()
	} else if gm.Config.GitHub.Auth.Token != "" {
		// TODO : implement token auth
		return nil, errors.New("token auth not implemented")
	} else {
		return nil, errors.New("failed to get auth method due to missing config")
	}
}

func (gm *GitManager) getAuthMethodPublicKey() (*ssh.PublicKeys, error) {
	var publicKey *ssh.PublicKeys
	sshPath := gm.Config.GitHub.Auth.SSHKeyPath
	sshKey, _ := os.ReadFile(sshPath)
	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if err != nil {
		return nil, err
	}
	return publicKey, err
}

// getRepoCloneDir() method is used to get the directory where a git repository
// will be cloned by this GitManager instance.
func (gm *GitManager) getRepoCloneDir(repoURL string) (string, error) {
	repoName, err := parseRepoNameFromURL(repoURL)
	if err != nil {
		return "", err
	}
	cloneDir := gm.Config.Command.WorkDir + "/" + repoName
	return cloneDir, nil
}
