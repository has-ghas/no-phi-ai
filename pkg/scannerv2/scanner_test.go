package scannerv2

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
)

var (
	test_empty_repository_map = make(map[string]*ScanRepository)
	test_context              = context.Background()
	test_failed_msg           = "failed test : %s"
	test_log_level            = "trace"
	test_work_dir             = "/tmp/no-phi-ai/test/pkg/scannerv2"
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
)

// TestNewScanner unit test function tests the NewScanner() function.
func TestNewScanner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config_func  func() *cfg.Config
		ctx          context.Context
		err_expected bool
		name         string
	}{
		{
			config_func:  func() *cfg.Config { return &cfg.Config{} },
			ctx:          test_context,
			err_expected: true,
			name:         "Config_Empty",
		},
		{
			config_func:  cfg.NewDefaultConfig,
			ctx:          test_context,
			err_expected: true,
			name:         "Config_Default",
		},
		{
			config_func: func() *cfg.Config {
				c := cfg.NewDefaultConfig()
				c.App.Log.Level = test_log_level
				// c.AzureAI.AuthKey = "test-auth-key" <- missing creates error
				c.AzureAI.DryRun = true
				c.AzureAI.Service = "test-service"
				c.Git.Auth.Token = "test-token"
				c.Git.WorkDir = test_work_dir
				return c
			},
			ctx:          test_context,
			err_expected: true,
			name:         "Config_Missing_AzureAIAuthKey",
		},
		{
			config_func:  test_valid_config_func,
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
				NewMemoryResultRecordIO(test_context),
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

// TestScanner_Run() unit test function tests the Run() method of a new Scanner.
func TestScanner_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config_func  func() *cfg.Config
		ctx          context.Context
		err_chan     chan error
		err_expected error
		name         string
		req_chan     chan<- Request
		resp_chan    <-chan Response
	}{
		{
			config_func:  test_valid_config_func,
			ctx:          test_context,
			err_chan:     make(chan error),
			err_expected: nil,
			name:         "Scanner_Run_Pass",
			req_chan:     make(chan<- Request),
			resp_chan:    make(<-chan Response),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := test.config_func()
			scanner, scanner_err := NewScanner(
				test.ctx,
				config,
				NewMemoryResultRecordIO(test_context),
			)
			if !assert.NoErrorf(t, scanner_err, test_failed_msg, test.name) {
				assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
			}

			go scanner.Run(test.err_chan, test.req_chan, test.resp_chan)
			err := <-test.err_chan

			if test.err_expected == nil {
				assert.NoErrorf(t, err, test_failed_msg, test.name)
			} else {
				assert.Equalf(t, test.err_expected, err, test_failed_msg, test.name)
			}
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
				test_valid_config_func(),
				NewMemoryResultRecordIO(test_context),
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
	scanner, scanner_err := NewScanner(test_context, test_valid_config_func(), NewMemoryResultRecordIO(test_context))
	if !assert.NoErrorf(t, scanner_err, test_failed_msg, "ProcessRequests") {
		assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
	}

	// create input and output channels
	chan_requests_in := make(chan Request)
	chan_requests_out := make(chan<- Request)
	chan_errors_out := make(chan error)

	// start the requests processor
	go scanner.processRequests(chan_requests_in, chan_requests_out, chan_errors_out)

	chan_requests_in <- Request{}
	err2 := <-chan_errors_out
	assert.Equal(t, ErrProcessRequestNoID, err2)

	// close the input channel to stop the requests processor
	close(chan_requests_in)

	// wait for the requests processor to finish
	time.Sleep(time.Millisecond) // Sleep for a short duration to allow goroutine to exit
}

// TestScanner_processResponses() unit test function tests the
// processResponses method of the Scanner object type.
func TestScanner_processResponses(t *testing.T) {
	t.Parallel()
	// create a new Scanner instance
	scanner, scanner_err := NewScanner(
		test_context,
		test_valid_config_func(),
		NewMemoryResultRecordIO(test_context),
	)
	if !assert.NoErrorf(t, scanner_err, test_failed_msg, "ProcessResponses") {
		assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
	}

	// create input and output channels
	chan_responses_in := make(chan Response)
	chan_errors_out := make(chan error)

	// start the response processor
	go scanner.processResponses(chan_responses_in, chan_errors_out)

	chan_responses_in <- NewResponse(&Request{})
	err2 := <-chan_errors_out
	assert.Equal(t, ErrProcessResponseNoID, err2)

	// close the input channel to stop the response processor
	close(chan_responses_in)

	// wait for the response processor to finish
	time.Sleep(time.Millisecond) // Sleep for a short duration to allow goroutine to exit
}

// TestScanner_scan() unit test function tests the scan() method of a new
// Scanner.
func TestScanner_scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config_func  func() *cfg.Config
		ctx          context.Context
		err_chan     chan error
		err_expected error
		name         string
	}{
		{
			config_func:  test_valid_config_func,
			ctx:          test_context,
			name:         "Scanner_Scan_Pass",
			err_chan:     make(chan error),
			err_expected: nil,
		},
		{
			config_func:  test_valid_config_func,
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
				NewMemoryResultRecordIO(test_context),
			)
			if !assert.NoErrorf(t, scanner_err, test_failed_msg, test.name) {
				assert.FailNowf(t, "failed to create scanner : %s", scanner_err.Error())
			}

			if test.err_chan == nil {
				assert.Panics(t, func() { scanner.scan(nil) })
				return
			}
			go scanner.scan(test.err_chan)
			err := <-test.err_chan

			if test.err_expected == nil {
				assert.NoErrorf(t, err, test_failed_msg, test.name)
			} else {
				assert.Equalf(t, test.err_expected, err, test_failed_msg, test.name)
			}
		})
	}
}
