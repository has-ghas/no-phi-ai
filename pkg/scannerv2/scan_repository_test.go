package scannerv2

import (
	"context"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/stretchr/testify/assert"
)

var (
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

// TestScanRepository_GetRepository tests the GetRepository() method
// of the ScanRepository struct.
func TestScanRepository_GetRepository(t *testing.T) {
	ctx := context.Background()
	channel_requests := make(chan<- Request)
	name := "TestRepositoryInit"

	t.Run(name, func(t *testing.T) {
		repo, err := NewScanRepository(ctx, test_repo_url, test_repo_scan_config, channel_requests)
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
