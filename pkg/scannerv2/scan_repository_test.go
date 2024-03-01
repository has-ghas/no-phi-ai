package scannerv2

import (
	"context"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/stretchr/testify/assert"

	"github.com/has-ghas/no-phi-ai/pkg/rrr"
)

var (
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
	t.Parallel()

	ctx := context.Background()
	channel_errors := make(chan<- error)
	channel_requests := make(chan<- rrr.Request)

	t.Run("ValidInput", func(t *testing.T) {
		repo, err := NewScanRepository(
			ctx,
			test_repo_url,
			test_repo_scan_config,
			channel_requests,
			channel_errors,
		)

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
		repo, err := NewScanRepository(
			nil,
			test_repo_url,
			test_repo_scan_config,
			channel_requests,
			channel_errors,
		)

		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.ErrorContains(t, err, ErrScanRepositoryContextNil.Error())
		assert.Nil(t, repo)
	})

	t.Run("NilChannelErrors", func(t *testing.T) {
		repo, err := NewScanRepository(
			ctx,
			test_repo_url,
			test_repo_scan_config,
			channel_requests,
			nil,
		)

		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.ErrorContains(t, err, ErrScanRepositoryChannelErrorsNil.Error())
		assert.Nil(t, repo)
	})

	t.Run("NilChannelRequests", func(t *testing.T) {
		repo, err := NewScanRepository(
			ctx,
			test_repo_url,
			test_repo_scan_config,
			nil,
			channel_errors,
		)

		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.ErrorContains(t, err, ErrScanRepositoryChannelRequestsNil.Error())
		assert.Nil(t, repo)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		invalidURL := "invalid-url"
		repo, err := NewScanRepository(
			ctx,
			invalidURL,
			test_repo_scan_config,
			channel_requests,
			channel_errors,
		)

		assert.ErrorContains(t, err, ErrMsgScanRepositoryCreate)
		assert.Nil(t, repo)
	})
}

// TestScanRepository_GetRepository tests the GetRepository() method
// of the ScanRepository struct.
func TestScanRepository_GetRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	channel_errors := make(chan<- error)
	channel_requests := make(chan<- rrr.Request)
	name := "TestRepositoryInit"

	t.Run(name, func(t *testing.T) {
		repo, err := NewScanRepository(
			ctx,
			test_repo_url,
			test_repo_scan_config,
			channel_requests,
			channel_errors,
		)
		if !assert.NoError(t, err) || !assert.NotNil(t, repo) {
			t.FailNow()
		}

		assert.NotEmpty(t, repo.ID)

		// confirm that the repository is nil before adding it to the
		// ScanRepository object
		result1 := repo.GetRepository()
		assert.Nil(t, result1)

		// initialize the bare *git.Repository and add it to the repo object
		var init_err error
		repo.repository, init_err = git.Init(memory.NewStorage(), nil)
		assert.NoError(t, init_err)

		// get the repository after adding it to the ScanRepository object
		result2 := repo.GetRepository()
		assert.NotNil(t, result2)
	})
}

// TestScanRepository_setRepository tests the GetRepository() method
// of the ScanRepository struct.
func TestScanRepository_setRepository(t *testing.T) {
	t.Parallel()

	channel_errors := make(chan<- error)
	channel_requests := make(chan<- rrr.Request)

	tests := []struct {
		name                 string
		expected_init_err    error
		expected_set_err     error
		repository_init_func func() (*git.Repository, error)
	}{
		{
			name:              "InvalidRepositoryPointer",
			expected_init_err: nil,
			expected_set_err:  ErrScanRepositorySetRepositoryNil,
			repository_init_func: func() (*git.Repository, error) {
				return nil, nil
			},
		},
		{
			name:              "ValidRepositoryInit",
			expected_init_err: nil,
			expected_set_err:  nil,
			repository_init_func: func() (*git.Repository, error) {
				return git.Init(memory.NewStorage(), nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewScanRepository(
				test_context,
				test_repo_url,
				test_repo_scan_config,
				channel_requests,
				channel_errors,
			)
			if !assert.NoError(t, err) || !assert.NotNil(t, repo) {
				t.FailNow()
			}
			assert.NotEmpty(t, repo.ID)

			// initialize the bare *git.Repository and add it to the repo object
			repository, init_err := test.repository_init_func()
			if test.expected_init_err == nil {
				assert.NoError(t, init_err)
			} else {
				assert.ErrorIs(t, init_err, test.expected_init_err)
			}

			set_err := repo.setRepository(repository)
			if test.expected_set_err == nil {
				assert.NoError(t, set_err)
			} else {
				assert.ErrorIs(t, set_err, test.expected_set_err)
			}

			// get the repository after adding it to the ScanRepository object
			result := repo.GetRepository()
			assert.Equal(t, repository, result)
		})
	}
}
