package scannerv2

import (
	"context"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
	"github.com/stretchr/testify/assert"
)

var (
	test_deploy_key_path  = "./testdata/deploy_key.test1"
	test_repo_work_dir    = "/tmp/no-phi-ai/test/pkg/scannerv2/scan_repository"
	test_repo_name        = ".test_NO-PHI-AI"
	test_repo_org         = "data-douser"
	test_repo_scan_config = &cfg.GitScanConfig{
		IgnoreRepositories: []string{},
		Repositories:       []string{test_repo_url},
	}
	test_repo_url = "git@github.com:" + test_repo_org + "/" + test_repo_name + ".git"
)

// TestNewScanRepository unit test function tests the NewScanRepository() function.
func TestNewScanRepository(t *testing.T) {
	ctx := context.Background()
	channel_requests := make(chan<- Request)

	t.Run("ValidInput", func(t *testing.T) {
		repo, err := NewScanRepository(ctx, test_repo_url, test_repo_scan_config, channel_requests)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.NotEmpty(t, repo.ID)
		assert.Equal(t, test_repo_url, repo.ID)
		assert.Equal(t, test_repo_name, repo.Name)
		assert.Equal(t, test_repo_url, repo.URL)
		assert.Equal(t, channel_requests, repo.channel_requests)
		assert.Equal(t, ctx, repo.ctx)
		assert.NotNil(t, repo.logger)
		assert.Nil(t, repo.repository)
		assert.NotNil(t, repo.TrackerCommits)
		assert.NotNil(t, repo.TrackerFiles)
	})

	t.Run("NilContext", func(t *testing.T) {
		repo, err := NewScanRepository(nil, test_repo_url, test_repo_scan_config, channel_requests)
		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.ErrorContains(t, err, ErrScanRepositoryContextNil.Error())
		assert.Nil(t, repo)
	})

	t.Run("NilChannelDocuments", func(t *testing.T) {
		repo, err := NewScanRepository(ctx, test_repo_url, test_repo_scan_config, nil)
		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.ErrorContains(t, err, ErrScanRepositoryChannelDocumentsNil.Error())
		assert.Nil(t, repo)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		invalidURL := "invalid-url"
		repo, err := NewScanRepository(ctx, invalidURL, test_repo_scan_config, channel_requests)
		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.Nil(t, repo)
	})
}

// TestScanRepository_clone_and_GetRepository tests the clone() and
// GetRepository() methods of the ScanRepository struct.
func TestScanRepository_clone_and_GetRepository(t *testing.T) {
	ctx := context.Background()
	channel_requests := make(chan<- Request)
	tests := []struct {
		expected_err_msg string
		expected_repo    *git.Repository
		git_config       *cfg.GitConfig
		name             string
		url              string
	}{
		{
			expected_repo:    &git.Repository{},
			expected_err_msg: "",
			git_config: &cfg.GitConfig{
				Auth: cfg.GitAuthConfig{
					SSHKeyPath: test_deploy_key_path,
				},
				WorkDir: test_repo_work_dir,
			},
			name: "ValidConfig",
			url:  test_repo_url,
		},
		{
			expected_repo:    nil,
			expected_err_msg: "failed to get auth method due to missing config",
			git_config: &cfg.GitConfig{
				Auth: cfg.GitAuthConfig{
					SSHKeyPath: "",
				},
				WorkDir: test_repo_work_dir,
			},
			name: "InvalidConfig",
			url:  test_repo_url,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewScanRepository(ctx, test.url, test_repo_scan_config, channel_requests)
			if !assert.NoError(t, err) || !assert.NotNil(t, repo) {
				t.FailNow()
			}

			assert.NotEmpty(t, repo.ID)

			var git_manager *nogit.GitManager
			if test.git_config != nil {
				git_manager = nogit.NewGitManager(test.git_config, ctx, repo.logger)
			} else {
				git_manager = nil
			}

			// clean the repository directory to force a fresh clone
			clean_err := git_manager.CleanRepo(test.url)
			if !assert.NoError(t, clean_err) {
				t.FailNow()
			}

			// clone the repository
			clone_err := repo.clone(git_manager)
			if clone_err != nil && test.expected_err_msg != "" {
				if !assert.ErrorContains(t, clone_err, test.expected_err_msg) {
					t.FailNow()
				}
				return
			}
			if !assert.NoError(t, clone_err) {
				t.FailNow()
			}

			result := repo.GetRepository()
			if test.expected_repo == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}

			// clean the repository directory to force a fresh clone
			post_clean_err := git_manager.CleanRepo(test.url)
			assert.NoError(t, post_clean_err)
		})
	}
}
