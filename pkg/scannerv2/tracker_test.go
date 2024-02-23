package scannerv2

import (
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var (
	test_message_complete = "test message complete"
	test_message_ignore   = "test message ignore"
	test_message_init     = "test message init"
	test_message_error    = "test message error"
	test_message_warning  = "test message warning"
)

// TestKeyCodeToState() unit test function is used to test the KeyCodeToState()
// function.
func TestKeyCodeToState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code     int
		expected string
		name     string
	}{
		{
			code:     KeyCodeComplete,
			expected: KeyStateComplete,
			name:     "KeyCodeComplete",
		},
		{
			code:     KeyCodeIgnore,
			expected: KeyStateIgnore,
			name:     "KeyCodeIgnore",
		},
		{
			code:     KeyCodeInit,
			expected: KeyStateInit,
			name:     "KeyCodeInit",
		},
		{
			code:     KeyCodeError,
			expected: KeyStateError,
			name:     "KeyCodeError",
		},
		{
			code:     KeyCodeWarning,
			expected: KeyStateWarning,
			name:     "KeyCodeWarning",
		},
		{
			code:     123, // Replace with your custom code value
			expected: KeyStateInit,
			name:     "CustomCode",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := KeyCodeToState(test.code)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestKeyCodeValidate() unit test function tests the KeyCodeValidate() function.
func TestKeyCodeValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code     int
		expected error
		name     string
	}{
		{
			code:     KeyCodeInit,
			expected: nil,
			name:     "KeyCodeInit",
		},
		{
			code:     KeyCodeError,
			expected: nil,
			name:     "KeyCodeError",
		},
		{
			code:     KeyCodeIgnore,
			expected: nil,
			name:     "KeyCodeIgnore",
		},
		{
			code:     KeyCodeWarning,
			expected: nil,
			name:     "KeyCodeWarning",
		},
		{
			code:     KeyCodeComplete,
			expected: nil,
			name:     "KeyCodeComplete",
		},
		{
			code:     KeyCodeInit - 1,
			expected: ErrKeyCodeInvalid,
			name:     "Code_Invalid_Low",
		},
		{
			code:     KeyCodeComplete + 1,
			expected: ErrKeyCodeInvalid,
			name:     "Code_Invalid_High",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := KeyCodeValidate(test.code)
			assert.Equal(t, test.expected, err)
		})
	}
}

// TestNewKeyData() unit test function tests the NewKeyData() function.
func TestNewKeyData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code             int
		expected_data    KeyData
		expected_err     error
		expected_message string
		message          string
		name             string
	}{
		{
			code: KeyCodeInit,
			expected_data: KeyData{
				Code:  KeyCodeInit,
				State: KeyStateInit,
			},
			expected_err:     nil,
			expected_message: test_message_init,
			message:          test_message_init,
			name:             "ValidCodeInit",
		},
		{
			code: KeyCodeInit,
			expected_data: KeyData{
				Code:  KeyCodeInit,
				State: KeyStateInit,
			},
			expected_err:     nil,
			expected_message: "",
			message:          "",
			name:             "ValidCodeInitMessageEmpty",
		},
		{
			code: KeyCodeError,
			expected_data: KeyData{
				Code:  KeyCodeError,
				State: KeyStateError,
			},
			expected_err:     nil,
			expected_message: test_message_error,
			message:          test_message_error,
			name:             "ValidCodeError",
		},
		{
			code: KeyCodeIgnore,
			expected_data: KeyData{
				Code:  KeyCodeIgnore,
				State: KeyStateIgnore,
			},
			expected_err:     nil,
			expected_message: test_message_ignore,
			message:          test_message_ignore,
			name:             "ValidCodeIgnore",
		},
		{
			code: KeyCodeWarning,
			expected_data: KeyData{
				Code:  KeyCodeWarning,
				State: KeyStateWarning,
			},
			expected_err:     nil,
			expected_message: test_message_warning,
			message:          test_message_warning,
			name:             "ValidCodeWarning",
		},
		{
			code: KeyCodeComplete,
			expected_data: KeyData{
				Code:  KeyCodeComplete,
				State: KeyStateComplete,
			},
			expected_err:     nil,
			expected_message: test_message_complete,
			message:          test_message_complete,
			name:             "ValidCodeComplete",
		},
		{
			code:             KeyCodeInit - 1,
			expected_data:    KeyData{},
			expected_err:     ErrKeyCodeInvalid,
			expected_message: "",
			message:          "",
			name:             "InvalidCodeLow",
		},
		{
			code:             KeyCodeComplete + 1,
			expected_data:    KeyData{},
			expected_err:     ErrKeyCodeInvalid,
			expected_message: "",
			message:          "",
			name:             "InvalidCodeHigh",
		},
	}

	timestamp_test_min := TimestampNow()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := NewKeyData(test.code, test.message)
			if test.expected_err != nil {
				assert.ErrorContains(t, err, test.expected_err.Error())
				return
			}
			timestamp_test_max := TimestampNow()
			assert.NoError(t, err)
			assert.Equal(t, test.expected_data.Code, data.Code)
			assert.Equal(t, test.expected_data.State, data.State)
			assert.Equal(t, test.expected_message, data.Message)
			assert.GreaterOrEqual(t, data.TimestampFirst, timestamp_test_min)
			assert.LessOrEqual(t, data.TimestampFirst, timestamp_test_max)
			assert.GreaterOrEqual(t, data.TimestampLatest, timestamp_test_min)
			assert.LessOrEqual(t, data.TimestampLatest, timestamp_test_max)
			assert.Exactly(t, data.TimestampFirst, data.TimestampLatest, "TimestampFirst and TimestampLatest should be equal for a new KeyData instance.")
		})
	}
}

// TestNewKeyDataCounts() unit test function tests the NewKeyDataCounts function.
func TestNewKeyDataCounts(t *testing.T) {
	t.Parallel()

	expectedCounts := KeyDataCounts{
		Complete: 0,
		Error:    0,
		Ignore:   0,
		Init:     0,
		Warning:  0,
	}

	counts := NewKeyDataCounts()

	assert.Equal(t, expectedCounts, counts)
}

// TestNewKeyTracker() unit test function tests the NewKeyTracker function.
func TestNewKeyTracker(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(os.Stdout)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	tests := []struct {
		kind         string
		expected     *KeyTracker
		expected_err error
		name         string
	}{
		{
			kind: ScanObjectTypeCommit,
			expected: &KeyTracker{
				keys:   make(map[string]KeyData, 0),
				kind:   ScanObjectTypeCommit,
				logger: &logger,
				mu:     &sync.RWMutex{},
			},
			expected_err: nil,
			name:         "ValidKindCommit",
		},
		{
			kind: ScanObjectTypeDocument,
			expected: &KeyTracker{
				keys:   make(map[string]KeyData, 0),
				kind:   ScanObjectTypeDocument,
				logger: &logger,
				mu:     &sync.RWMutex{},
			},
			expected_err: nil,
			name:         "ValidKindDocument",
		},
		{
			kind: ScanObjectTypeFile,
			expected: &KeyTracker{
				keys:   make(map[string]KeyData, 0),
				kind:   ScanObjectTypeFile,
				logger: &logger,
				mu:     &sync.RWMutex{},
			},
			expected_err: nil,
			name:         "ValidKindFile",
		},
		{
			kind: ScanObjectTypeRepository,
			expected: &KeyTracker{
				keys:   make(map[string]KeyData, 0),
				kind:   ScanObjectTypeRepository,
				logger: &logger,
				mu:     &sync.RWMutex{},
			},
			expected_err: nil,
			name:         "ValidKindRepository",
		},
		{
			kind:         "InvalidKind",
			expected:     nil,
			expected_err: ErrKeyTrackerInvalidKind,
			name:         "InvalidKind",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracker, err := NewKeyTracker(test.kind, &logger)
			if test.expected_err != nil {
				assert.ErrorIs(t, err, test.expected_err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected.keys, tracker.keys)
			assert.Equal(t, test.expected.kind, tracker.kind)
			assert.Equal(t, test.expected.logger, tracker.logger)
			assert.Equal(t, test.expected.mu, tracker.mu)
		})
	}
}

// TestKeyTracker_Get() unit test function tests the Get() method of the
// KeyTracker type.
func TestKeyTracker_Get(t *testing.T) {
	t.Parallel()

	// Create a new KeyTracker instance
	tracker := &KeyTracker{
		keys:   map[string]KeyData{},
		kind:   ScanObjectTypeFile,
		logger: nil,
		mu:     &sync.RWMutex{},
	}

	// add some test data to the tracker
	testKey := "testKey"
	testData := KeyData{
		Code:  KeyCodeInit,
		State: KeyStateInit,
	}
	tracker.keys[testKey] = testData

	// Test case: Existing key
	t.Run("ExistingKey", func(t *testing.T) {
		keyData, exists := tracker.Get(testKey)
		assert.True(t, exists)
		assert.Equal(t, testData, keyData)
	})

	// Test case: Non-existing key
	t.Run("NonExistingKey", func(t *testing.T) {
		keyData, exists := tracker.Get("nonExistingKey")
		assert.False(t, exists)
		assert.Equal(t, KeyData{}, keyData)
	})

	// Test case: Empty key
	t.Run("EmptyKey", func(t *testing.T) {
		keyData, exists := tracker.Get("")
		assert.False(t, exists)
		assert.Equal(t, KeyData{}, keyData)
	})
}

// TestKeyTracker_GetCounts() unit test function tests the GetCounts()
// method of the KeyTracker type.
func TestKeyTracker_GetCounts(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(os.Stdout)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	type test_data struct {
		key   string
		codes []int
	}

	tests := []struct {
		name      string
		data      []test_data
		final     KeyDataCounts
		final_err error
		kind      string
	}{
		{
			name:      "InvalidKind",
			data:      []test_data{},
			final:     KeyDataCounts{},
			final_err: ErrKeyTrackerInvalidKind,
			kind:      "InvalidKind",
		},
		{
			name: "Init_1",
			data: []test_data{
				{
					codes: []int{KeyCodeInit},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 0,
				Error:    0,
				Ignore:   0,
				Init:     1,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "Init_2",
			data: []test_data{
				{
					codes: []int{KeyCodeInit},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 0,
				Error:    0,
				Ignore:   0,
				Init:     1,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "Complete_1",
			data: []test_data{
				{
					codes: []int{KeyCodeComplete},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 1,
				Error:    0,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "Complete_2",
			data: []test_data{
				{
					codes: []int{KeyCodeComplete, 2},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 1,
				Error:    0,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "Error_1",
			data: []test_data{
				{
					codes: []int{KeyCodeError},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 0,
				Error:    1,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "Error_2",
			data: []test_data{
				{
					codes: []int{KeyCodeError, -1},
					key:   "A",
				},
			},
			final: KeyDataCounts{
				Complete: 0,
				Error:    1,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "Mixed_1",
			data: []test_data{
				{
					codes: []int{KeyCodeInit},
					key:   "A",
				},
				{
					codes: []int{KeyCodeInit, KeyCodeError, KeyCodeInit},
					key:   "B",
				},
				{
					codes: []int{KeyCodeIgnore},
					key:   "C",
				},
				{
					codes: []int{KeyCodeInit, KeyCodeWarning, KeyCodeError},
					key:   "D",
				},
				{
					codes: []int{KeyCodeInit, KeyCodeError, KeyCodeComplete},
					key:   "E",
				},
			},
			final: KeyDataCounts{
				Complete: 1,
				Error:    1,
				Ignore:   1,
				Init:     1,
				Warning:  1,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "Progression",
			data: []test_data{
				{
					codes: []int{-2, -1, 0, 1, 2},
					key:   "A",
				},
				{
					codes: []int{
						KeyCodeInit,
						KeyCodeError,
						KeyCodeIgnore,
						KeyCodeWarning,
						KeyCodeComplete,
					},
					key: "B",
				},
			},
			final: KeyDataCounts{
				Complete: 2,
				Error:    0,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeFile,
		},
		{
			name: "Regression",
			data: []test_data{
				{
					codes: []int{2, 1, 0, -1, -2},
					key:   "A",
				},
				{
					codes: []int{
						KeyCodeComplete,
						KeyCodeWarning,
						KeyCodeIgnore,
						KeyCodeError,
						KeyCodeInit,
					},
					key: "B",
				},
			},
			final: KeyDataCounts{
				Complete: 2,
				Error:    0,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "ReInit",
			data: []test_data{
				{
					codes: []int{
						-2,
						-1,
						0,
						1,
						2,
						-2,
						-2,
						-2,
					},
					key: "A",
				},
				{
					codes: []int{
						KeyCodeInit,
						KeyCodeError,
						KeyCodeIgnore,
						KeyCodeWarning,
						KeyCodeComplete,
						KeyCodeInit,
						KeyCodeInit,
						KeyCodeInit,
					},
					key: "B",
				},
			},
			final: KeyDataCounts{
				Complete: 2,
				Error:    0,
				Ignore:   0,
				Init:     0,
				Warning:  0,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
	}

	for _, test_i := range tests {
		t.Run(test_i.name, func(t *testing.T) {
			tracker, err := NewKeyTracker(test_i.kind, &logger)
			if test_i.final_err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, test_i.final_err)
				return
			}

			// for each key in the test data, apply the series of update codes
			for _, d := range test_i.data {
				for _, code := range d.codes {
					_, update_err := tracker.Update(d.key, code, "")
					assert.NoError(t, update_err)
				}
			}

			// get the data for the key after applying all updates
			key_data_counts := tracker.GetCounts()
			assert.Equal(t, test_i.final, key_data_counts)
		})
	}
}

// TestKeyTracker_GetKeys unit test function tests the GetKeys() method of the KeyTracker type.
func TestKeyTracker_GetKeys(t *testing.T) {
	t.Parallel()

	// create a new KeyTracker instance
	tracker := &KeyTracker{
		keys:   map[string]KeyData{},
		kind:   ScanObjectTypeFile,
		logger: nil,
		mu:     &sync.RWMutex{},
	}

	// add some test data to the tracker
	tracker.keys["key1"] = KeyData{
		Code:  KeyCodeInit,
		State: KeyStateInit,
	}
	tracker.keys["key2"] = KeyData{
		Code:  KeyCodeComplete,
		State: KeyStateComplete,
	}
	tracker.keys["key3"] = KeyData{
		Code:  KeyCodeError,
		State: KeyStateError,
	}

	// call the GetKeys() method
	keys := tracker.GetKeys()

	// assert the expected keys
	expectedKeys := []string{"key1", "key2", "key3"}
	assert.ElementsMatch(t, expectedKeys, keys)
}

// TestKeyTracker_GetKeysData unit test function tests the GetKeysData() method of the KeyTracker type.
func TestKeyTracker_GetKeysData(t *testing.T) {
	t.Parallel()

	// create a new KeyTracker instance
	tracker := &KeyTracker{
		keys:   map[string]KeyData{},
		kind:   ScanObjectTypeFile,
		logger: nil,
		mu:     &sync.RWMutex{},
	}

	// add some test data to the tracker
	test_key_1 := "test_key_1"
	test_data_1 := KeyData{
		Code:  KeyCodeInit,
		State: KeyStateInit,
	}
	tracker.keys[test_key_1] = test_data_1
	test_key_2 := "test_key_2"
	test_data_2 := KeyData{
		Code:    KeyCodeComplete,
		Message: test_message_complete,
		State:   KeyStateComplete,
	}
	tracker.keys[test_key_2] = test_data_2
	test_key_3 := "test_key_3"
	test_data_3 := KeyData{
		Code:    KeyCodeError,
		Message: test_message_error,
		State:   KeyStateError,
	}
	tracker.keys[test_key_3] = test_data_3

	// call the GetKeysData method
	keysData := tracker.GetKeysData()

	// check if the returned map is equal to the original keys map
	assert.Equal(t, tracker.keys, keysData)
}

// TestKeyTracker_PrintCodes unit test function tests the PrintCodes() method
// of the KeyTracker type.
func TestKeyTracker_PrintCodes(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(os.Stdout)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// create a new KeyTracker instance
	tracker := &KeyTracker{
		keys:   map[string]KeyData{},
		kind:   ScanObjectTypeFile,
		logger: &logger,
		mu:     &sync.RWMutex{},
	}

	// add some test data to the tracker
	tracker.keys["key1"] = KeyData{
		Code:  KeyCodeInit,
		State: KeyStateInit,
	}
	tracker.keys["key2"] = KeyData{
		Code:  KeyCodeComplete,
		State: KeyStateComplete,
	}
	tracker.keys["key3"] = KeyData{
		Code:  KeyCodeError,
		State: KeyStateError,
	}

	expected_codes := []int{KeyCodeInit, KeyCodeComplete, KeyCodeError}

	// call the PrintCodes() methodV
	actual_codes := tracker.PrintCodes()
	for _, code := range actual_codes {
		//assert.Equal(t, expected_codes, actual_codes)
		assert.Contains(t, expected_codes, code)
	}
}

// TestKeyTracker_PrintCounts unit test function tests the PrintCounts() method
// of the KeyTracker type.
func TestKeyTracker_PrintCounts(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(os.Stdout)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// create a new KeyTracker instance
	tracker := &KeyTracker{
		keys:   map[string]KeyData{},
		kind:   ScanObjectTypeFile,
		logger: &logger,
		mu:     &sync.RWMutex{},
	}

	// add some test data to the tracker
	tracker.keys["key1"] = KeyData{
		Code:  KeyCodeInit,
		State: KeyStateInit,
	}
	tracker.keys["key2"] = KeyData{
		Code:  KeyCodeComplete,
		State: KeyStateComplete,
	}
	tracker.keys["key3"] = KeyData{
		Code:  KeyCodeError,
		State: KeyStateError,
	}

	expected_key_data_counts := KeyDataCounts{
		Complete: 1,
		Error:    1,
		Ignore:   0,
		Init:     1,
		Warning:  0,
	}

	// call the PrintCounts() methodV
	actual_key_data_counts := tracker.PrintCounts()
	assert.Equal(t, expected_key_data_counts, actual_key_data_counts)
}

// TestKeyTracker_Update() unit test function tests the Update() method
// of the KeyTracker type.
func TestKeyTracker_Update(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(os.Stdout)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	type test_data struct {
		code        int
		expect_code int
		message     string
	}

	tests := []struct {
		name      string
		data      []test_data
		final     KeyData
		final_err error
		kind      string
	}{
		{
			name:      "InvalidKind",
			data:      []test_data{},
			final:     KeyData{},
			final_err: ErrKeyTrackerInvalidKind,
			kind:      "InvalidKind",
		},
		{
			name: "CodeInit",
			data: []test_data{
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeInit,
					message:     "",
				},
			},
			final: KeyData{
				Code:    KeyCodeInit,
				Message: "",
				State:   KeyStateInit,
			},
			final_err: nil,
			kind:      ScanObjectTypeCommit,
		},
		{
			name: "CodeComplete",
			data: []test_data{
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeInit,
					message:     "",
				},
				{
					code:        KeyCodeComplete,
					expect_code: KeyCodeComplete,
					message:     "",
				},
			},
			final: KeyData{
				Code:    KeyCodeComplete,
				Message: "",
				State:   KeyStateComplete,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "CodeError",
			data: []test_data{
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeInit,
					message:     "",
				},
				{
					code:        KeyCodeError,
					expect_code: KeyCodeError,
					message:     test_message_error,
				},
			},
			final: KeyData{
				Code:    KeyCodeError,
				Message: test_message_error,
				State:   KeyStateError,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "Progression",
			data: []test_data{
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeInit,
					message:     "",
				},
				{
					code:        KeyCodeError,
					expect_code: KeyCodeError,
					message:     test_message_error,
				},
				{
					code:        KeyCodeIgnore,
					expect_code: KeyCodeIgnore,
					message:     test_message_ignore,
				},
				{
					code:        KeyCodeWarning,
					expect_code: KeyCodeWarning,
					message:     test_message_warning,
				},
				{
					code:        KeyCodeComplete,
					expect_code: KeyCodeComplete,
					message:     test_message_complete,
				},
			},
			final: KeyData{
				Code:    KeyCodeComplete,
				Message: test_message_complete,
				State:   KeyStateComplete,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "Regression",
			data: []test_data{
				{
					code:        KeyCodeComplete,
					expect_code: KeyCodeComplete,
					message:     test_message_complete,
				},
				{
					code:        KeyCodeWarning,
					expect_code: KeyCodeComplete,
					message:     test_message_warning,
				},
				{
					code:        KeyCodeIgnore,
					expect_code: KeyCodeComplete,
					message:     test_message_ignore,
				},
				{
					code:        KeyCodeError,
					expect_code: KeyCodeComplete,
					message:     test_message_error,
				},
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeComplete,
					message:     test_message_init,
				},
			},
			final: KeyData{
				Code:    KeyCodeComplete,
				Message: test_message_complete,
				State:   KeyStateComplete,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
		{
			name: "ReInit",
			data: []test_data{
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeInit,
					message:     "",
				},
				{
					code:        KeyCodeComplete,
					expect_code: KeyCodeComplete,
					message:     test_message_complete,
				},
				{
					code:        KeyCodeInit,
					expect_code: KeyCodeComplete,
					message:     test_message_init,
				},
			},
			final: KeyData{
				Code:    KeyCodeComplete,
				Message: test_message_complete,
				State:   KeyStateComplete,
			},
			final_err: nil,
			kind:      ScanObjectTypeDocument,
		},
	}

	for _, test_i := range tests {
		t.Run(test_i.name, func(t *testing.T) {
			tracker, err := NewKeyTracker(test_i.kind, &logger)
			if test_i.final_err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, test_i.final_err)
				return
			}

			for _, d := range test_i.data {
				updated_code, update_err := tracker.Update(test_i.name, d.code, d.message)
				assert.NoError(t, update_err)
				assert.Exactly(t, d.expect_code, updated_code)
			}

			// get the data for the key after applying all updates
			key_data, key_exists := tracker.Get(test_i.name)
			if !assert.True(t, key_exists) {
				t.FailNow()
			}
			assert.Equal(t, test_i.final.Code, key_data.Code)
			assert.Equal(t, test_i.final.Message, key_data.Message)
			assert.Equal(t, test_i.final.State, key_data.State)
		})
	}
}