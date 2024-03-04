package scanner

import (
	"context"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	gitmemory "github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/scanner/memory"
	"github.com/has-ghas/no-phi-ai/pkg/scanner/rrr"
	"github.com/has-ghas/no-phi-ai/pkg/scanner/tracker"
)

var (
	test_empty_repository_map = make(map[string]*ScanRepository)
	test_context              = context.Background()
	test_failed_msg           = "failed test : %s"
	test_log_level            = "trace"
	test_work_dir             = "/tmp/no-phi-ai/test/pkg/scanner"
	test_valid_config_func    = func() *cfg.Config {
		c := cfg.NewDefaultConfig()
		c.App.Log.Level = test_log_level
		c.AzureAI.AuthKey = "test-auth-key"
		c.AzureAI.DryRun = true
		c.AzureAI.Service = "test-service"
		c.Git.Auth.Token = "test-token"
		c.Git.WorkDir = test_work_dir
		return c
	}
	test_valid_git_config_func = func() *cfg.GitConfig {
		c := test_valid_config_func()
		return &c.Git
	}
)

// TestNewScanner unit test function tests the NewScanner() function.
func TestNewScanner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config_func  func() *cfg.GitConfig
		ctx          context.Context
		err_expected bool
		name         string
	}{
		{
			config_func: func() *cfg.GitConfig {
				return &cfg.GitConfig{}
			},
			ctx:          test_context,
			err_expected: false,
			name:         "Config_Empty",
		},
		{
			config_func: func() *cfg.GitConfig {
				c := cfg.NewDefaultConfig()
				return &c.Git
			},
			ctx:          test_context,
			err_expected: false,
			name:         "Config_Default",
		},
		{
			config_func: func() *cfg.GitConfig {
				c := cfg.NewDefaultConfig()
				c.Git.Auth.Token = "test-token"
				c.Git.WorkDir = test_work_dir
				return &c.Git
			},
			ctx:          test_context,
			err_expected: false,
			name:         "Config_Missing_AzureAIAuthKey",
		},
		{
			config_func:  test_valid_git_config_func,
			ctx:          test_context,
			err_expected: false,
			name:         "Config_Valid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := test.config_func()
			scanner, err := NewScanner(
				test.ctx,
				config,
				memory.NewMemoryResultRecordIO(test_context),
			)

			if test.err_expected {
				assert.Errorf(t, err, test_failed_msg, test.name)
				assert.Nilf(t, scanner, test_failed_msg, test.name)
			} else {
				assert.NoErrorf(t, err, test_failed_msg, test.name)
				if assert.NotNilf(t, scanner, test_failed_msg, test.name) {
					assert.NotEqualf(t, "", scanner.ID, test_failed_msg, test.name)
				}
			}
		})
	}
}

// TestScanner_Scan() unit test function tests the Scan() method of a new Scanner.
func TestScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config_func  func() *cfg.GitConfig
		ctx          context.Context
		err_chan     chan error
		err_expected error
		name         string
		req_chan     chan<- rrr.Request
		resp_chan    <-chan rrr.Response
	}{
		{
			config_func: func() *cfg.GitConfig {
				config := test_valid_config_func()
				config.Git.Scan.Repositories = []string{"test_repo_url_fail"}
				return &config.Git
			},
			ctx:          test_context,
			err_chan:     make(chan error),
			err_expected: nil,
			name:         "Scanner_Run_Pass_1",
			req_chan:     make(chan<- rrr.Request),
			resp_chan:    make(<-chan rrr.Response),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := test.config_func()
			scanner, scanner_err := NewScanner(
				test.ctx,
				config,
				memory.NewMemoryResultRecordIO(test_context),
			)
			if !assert.NoErrorf(t, scanner_err, test_failed_msg, test.name) {
				assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
			}

			go scanner.Scan(test.err_chan, test.req_chan, test.resp_chan)
			if test.err_expected != nil {
				err := <-test.err_chan
				assert.ErrorContainsf(t, err, test.err_expected.Error(), test_failed_msg, test.name)
			}

			// TODO
		})
	}
}

// TestScanner_addScanRepository tests the addScanRepository method of the Scanner
// object type.
func TestScanner_addScanRepository(t *testing.T) {
	test_expected_map_func_empty := func() map[string]*ScanRepository {
		return test_empty_repository_map
	}
	t.Parallel()
	tests := []struct {
		name              string
		repo              *ScanRepository
		expected_err      error
		expected_map_func func() map[string]*ScanRepository
	}{
		{
			name: "TestRepoValid",
			repo: &ScanRepository{
				ID:   test_repo_url,
				Name: test_repo_name,
				URL:  test_repo_url,
			},
			expected_err: nil,
			expected_map_func: func() map[string]*ScanRepository {
				out := make(map[string]*ScanRepository)
				out[test_repo_url] = &ScanRepository{
					ID:   test_repo_url,
					Name: test_repo_name,
					URL:  test_repo_url,
				}
				return out
			},
		},
		{
			name:              "NilRepoError",
			repo:              nil,
			expected_err:      errors.Wrap(ErrScannerAddScanRepositoryNil, ErrMsgAddScanRepository),
			expected_map_func: test_expected_map_func_empty,
		},
		{
			name: "RepoNameError",
			repo: &ScanRepository{
				ID:   "",
				Name: test_repo_name,
				URL:  test_repo_url,
			},
			expected_err:      errors.Wrap(ErrScannerAddScanRepositoryEmptyID, ErrMsgAddScanRepository),
			expected_map_func: test_expected_map_func_empty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner, err := NewScanner(
				test_context,
				test_valid_git_config_func(),
				memory.NewMemoryResultRecordIO(test_context),
			)
			if !assert.NoError(t, err) {
				assert.FailNow(t, "failed to create scanner")
			}

			err = scanner.addScanRepository(test.repo)

			// validate the expected error
			if test.expected_err == nil {
				assert.NoError(t, err)
				// test the getScanRepository() method
				get_repo, get_err := scanner.getScanRepository(test.repo.ID)
				assert.NoError(t, get_err)
				assert.Equal(t, test.repo, get_repo)
			} else {
				assert.Equal(t, test.expected_err.Error(), err.Error())
			}
			// validate the expected map of repositories
			assert.Equal(t, test.expected_map_func(), scanner.scan_repositories)
		})
	}
}

// TestScanner_processRequests() unit test function tests the
// processRequests method of the Scanner object type.
func TestScanner_processRequests(t *testing.T) {
	t.Parallel()
	// create a new Scanner instance
	scanner, scanner_err := NewScanner(
		test_context,
		test_valid_git_config_func(),
		memory.NewMemoryResultRecordIO(test_context),
	)
	if !assert.NoErrorf(t, scanner_err, test_failed_msg, "ProcessRequests") {
		assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
	}

	// create input and output channels
	chan_quit_in := make(chan struct{})
	chan_requests_in := make(chan rrr.Request)
	chan_requests_out := make(chan<- rrr.Request)
	chan_errors_out := make(chan error)

	// start the requests processor
	go scanner.processRequests(chan_quit_in, chan_requests_in, chan_requests_out, chan_errors_out)

	chan_requests_in <- rrr.Request{}
	err2 := <-chan_errors_out
	assert.Equal(t, ErrProcessRequestNoID, err2)

	// close the input channels to stop goroutines
	close(chan_requests_in)
	close(chan_quit_in)

	// wait for the requests processor to finish
	time.Sleep(time.Millisecond) // Sleep for a short duration to allow goroutine to exit
}

// TestScanner_processResponse unit test function tests the processResponse method of the Scanner object type.
func TestScanner_processResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr  error
		name         string
		responseFunc func() rrr.Response
	}{
		{
			expectedErr: nil,
			name:        "Scanner_processResponse_Pass_1",
			responseFunc: func() rrr.Response {
				request, request_err := rrr.NewRequest(test_repo_url, "commit_id", "object_id", "test_text_example")
				if !assert.NoError(t, request_err) {
					assert.FailNow(t, "failed to create test request and response")
				}
				response := rrr.NewResponse(&request)
				return response
			},
		},
		{
			expectedErr: ErrProcessResponseNoID,
			name:        "Scanner_processResponse_Fail_1",
			responseFunc: func() rrr.Response {
				request, request_err := rrr.NewRequest(test_repo_url, "commit_id", "object_id", "test_text_example")
				if !assert.NoError(t, request_err) {
					assert.FailNow(t, "failed to create test request and response")
				}
				response := rrr.NewResponse(&request)
				// delete the response ID
				response.ID = ""
				return response
			},
		},
	}

	scanner, scanner_err := NewScanner(
		test_context,
		test_valid_git_config_func(),
		memory.NewMemoryResultRecordIO(test_context),
	)
	if !assert.NoError(t, scanner_err) {
		assert.FailNow(t, "failed to create scanner")
	}

	// initialize the bare *git.Repository
	repository, init_err := git.Init(gitmemory.NewStorage(), nil)
	assert.NoError(t, init_err)

	scan_repo, err := NewScanRepository(NewScanRepositoryInput{
		ChannelErrors:   make(chan<- error),
		ChannelRequests: make(chan<- rrr.Request),
		Config:          test_repo_scan_config,
		Context:         test_context,
		Repository:      repository,
		URL:             test_repo_url,
	})
	// is_scan_complete must be set in order to ensure that the
	// processResponse method does not block indefinitely
	scan_repo.is_scan_complete = true

	if !assert.NoError(t, err) || !assert.NotNil(t, scan_repo) {
		assert.FailNow(t, "failed to create test repository for Scanner")
	}

	if add_err := scanner.addScanRepository(scan_repo); !assert.NoError(t, add_err) {
		assert.FailNow(t, "failed to add test repository to Scanner")
	}
	if !assert.NotEmpty(t, scan_repo.ID) {
		assert.FailNow(t, "failed to create test repository with non-empty ID")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			chan_errors_out := make(chan error)

			response := test.responseFunc()
			t.Log("response.ID:", response.ID)
			go scanner.processResponse(response, chan_errors_out)

			if test.expectedErr != nil {
				err := <-chan_errors_out
				assert.EqualError(t, err, test.expectedErr.Error())
				return
			}

			// sleep for a short duration to allow the response to be processed
			time.Sleep(time.Millisecond * 100)

			req_key_data, req_key_exists := scanner.TrackerRequests.Get(response.ID)
			assert.Truef(t, req_key_exists, "failed to find response ID in requests tracker : ID=%s", response.ID)
			assert.Equal(t, req_key_data.Code, tracker.KeyCodeComplete)

			test_scan_repo, test_scan_repo_err := scanner.getScanRepository(scan_repo.ID)
			if !assert.NoError(t, test_scan_repo_err) {
				assert.FailNow(t, "failed to get test repository from scanner")
			}

			commit_key_data, commit_key_exists := test_scan_repo.TrackerCommits.Get(response.Commit.ID)
			assert.Truef(t, commit_key_exists, "failed to find commit ID in commits tracker : ID=%s", response.Commit.ID)
			assert.Contains(t, commit_key_data.Children, response.Object.ID)
			assert.Equal(t, commit_key_data.Code, tracker.KeyCodeComplete)
			assert.Equal(t, commit_key_data.State, tracker.KeyStateComplete)

			file_key_data, file_key_exists := test_scan_repo.TrackerFiles.Get(response.Object.ID)
			assert.Truef(t, file_key_exists, "failed to find file ID in files tracker : ID=%s", response.Object.ID)
			assert.Contains(t, file_key_data.Children, response.ID)
			assert.Equal(t, file_key_data.Code, tracker.KeyCodeComplete)
			assert.Equal(t, file_key_data.State, tracker.KeyStateComplete)
		})
	}
}

// TestScanner_processResponses() unit test function tests the
// processResponses() method of the Scanner object type.
func TestScanner_processResponses(t *testing.T) {
	t.Parallel()
	// create a new Scanner instance
	scanner, scanner_err := NewScanner(
		test_context,
		test_valid_git_config_func(),
		memory.NewMemoryResultRecordIO(test_context),
	)
	if !assert.NoErrorf(t, scanner_err, test_failed_msg, "ProcessResponses") {
		assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
	}

	// create input and output channels
	chan_quit := make(chan struct{})
	chan_responses_in := make(chan rrr.Response)
	chan_errors_out := make(chan error)

	// start the response processor
	go scanner.processResponses(chan_quit, chan_responses_in, chan_errors_out)

	chan_responses_in <- rrr.NewResponse(&rrr.Request{})
	err2 := <-chan_errors_out
	assert.Equal(t, ErrProcessResponseNoID, err2)

	// close the input channel to stop the response processor
	close(chan_responses_in)

	// wait for the response processor to finish
	time.Sleep(time.Millisecond) // Sleep for a short duration to allow goroutine to exit
}

// TestScanner_scanRepository() unit test function tests the scanRepository()
// method of a new Scanner.
func TestScanner_scanRepository(t *testing.T) {
	t.Parallel()

	// initialize the bare *git.Repository
	repository, init_err := git.Init(gitmemory.NewStorage(), nil)
	assert.NoError(t, init_err)

	tests := []struct {
		config_func  func() *cfg.GitConfig
		ctx          context.Context
		err_chan     chan error
		err_expected error
		name         string
	}{
		{
			config_func:  test_valid_git_config_func,
			ctx:          test_context,
			name:         "Scanner_Scan_Fail_Repo_URL",
			err_chan:     make(chan error),
			err_expected: errors.New("failed to parse repo name : invalid path in URL"),
		},
		{
			config_func:  test_valid_git_config_func,
			ctx:          test_context,
			name:         "Scanner_Scan_Panic_Channel_Nil",
			err_chan:     nil,
			err_expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := test.config_func()
			scanner, scanner_err := NewScanner(
				test.ctx,
				config,
				memory.NewMemoryResultRecordIO(test_context),
			)
			if !assert.NoErrorf(t, scanner_err, test_failed_msg, test.name) {
				assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
			}

			if test.err_chan == nil {
				assert.Panics(t, func() {
					scanner.scanRepository("test_repo_url", repository, nil, make(chan<- struct{}))
				})
				return
			}
			go scanner.scanRepository("test_repo_url", repository, test.err_chan, make(chan<- struct{}))
			err := <-test.err_chan

			if test.err_expected == nil {
				assert.NoErrorf(t, err, test_failed_msg, test.name)
			} else {
				assert.ErrorContainsf(t, err, test.err_expected.Error(), test_failed_msg, test.name)
			}
		})
	}
}
