package scanner

import "time"

// Status struct is used to track the status of a scan for the
// associated ScanObject, where Status is embedded in ScanObject.
type Status struct {
	Code          int    `json:"code"`
	CompletedAt   int64  `json:"completed_at"`
	ErroredAt     int64  `json:"errored_at"`
	IgnoredAt     int64  `json:"ignored_at"`
	RequestedAt   int64  `json:"request_at"`
	RespondedAt   int64  `json:"response_at"`
	ResultCode    int    `json:"response_code"`
	ResultMessage string `json:"response_message"`
	StartedAt     int64  `json:"started_at"`
	State         string `json:"state"`
	StateMessage  string `json:"state_message"`
}

// NewStatus() function initializes a new Status struct with
// default values, meaning that the state message is empty,
// the state is set to "init", and timestamps are set to 0.
func NewStatus() *Status {
	return &Status{
		Code:          StatusInitCode,
		CompletedAt:   0,
		ErroredAt:     0,
		IgnoredAt:     0,
		RequestedAt:   0,
		RespondedAt:   0,
		ResultCode:    ResultCleanCode,
		ResultMessage: ResultInitMsg,
		StartedAt:     0,
		State:         StatusInitState,
		StateMessage:  "",
	}
}

// IsResultClean() method returns true if the result of the scan for the
// associated object has been marked as clean, meaning that the object's
// text has been scanned for PHI/PII "entities" and _NONE_ were found.
func (s *Status) IsResultClean() bool {
	return s.ResultCode == ResultCleanCode
}

// IsResultDirty() method returns true if the result of the scan for the
// associated object has been marked as dirty, meaning that the object's
// text has been scanned for PHI/PII "entities" and _SOME_ were found.
func (s *Status) IsResultDirty() bool {
	return s.ResultCode == ResultDirtyCode
}

// IsResultError() method returns true if the result of the scan for the
// associated object has been marked as dirty, meaning that the object's
// text was scanned but an error was encountered in the request/response
func (s *Status) IsResultError() bool {
	return s.ResultCode == ResultErrorCode
}

// IsResultUnknown() method returns true if the result of the scan for
// the associated object has been marked as unknown, which is the
// default/init value for the result code.
func (s *Status) IsResultUnknown() bool {
	return s.ResultCode == ResultInitCode
}

// IsCompleted() method returns true if the status of the scan for the
// associated object has been marked as completed, meaning that the
// object has been scanned along with its chidlren and/or text.
func (s *Status) IsCompleted() bool {
	return s.Code == StatusCompleteCode && s.CompletedAt > 0
}

// IsCompleted() method returns true if the status of the scan for the
// associated object has been marked as errored, meaning that some
// error was encountered while attempting to scan of the associated
// object.
func (s *Status) IsErrored() bool {
	return s.Code == StatusErrorCode && s.ErroredAt > 0
}

// IsIgnored() method returns true if the status of the scan for the
// associated object has been marked as ignored, meaning that the
// object's text will not be scanned but the object will still be
// tracked (as ignored) in the scan results.
func (s *Status) IsIgnored() bool {
	return s.Code == StatusIgnoredCode && s.IgnoredAt > 0
}

// IsRequested() method returns true if the status of the scan for the
// associated object has been marked as requested, meaning that a
// request has been sent to an external service scan the text of the
// object.
func (s *Status) IsRequested() bool {
	return s.Code == StatusProcessingRequestCode && s.RequestedAt > 0
}

// IsResponded() method returns true if the status of the scan for the
// associated object has been marked as responded, meaning that an
// external service responded to our request to scan the text of the
// object.
func (s *Status) IsResponded() bool {
	return s.Code == StatusProcessingResponseCode && s.RespondedAt > 0
}

// IsStarted() method returns true if the status of the scan for the
// associated object has been marked as started, meaning that the
// object was not ignored and the scan process has begun.
func (s *Status) IsStarted() bool {
	return s.Code == StatusStartCode && s.StartedAt > 0
}

// SetCompleted() method sets the appropriate fields to indicate that
// the code, message (optional), state, and timestamp for the completed
// scan of the associated object.
func (s *Status) SetCompleted(result_code int, state_msg string) {
	s.Code = StatusCompleteCode
	s.CompletedAt = time.Now().Unix()
	s.ErroredAt = 0
	s.ResultCode = result_code
	s.ResultMessage = GetMessageFromResultCode(result_code)
	s.State = StatusCompleteState
	if state_msg != "" {
		s.StateMessage = state_msg
	}
}

// SetErrored() method sets the appropriate fields to indicate that an
// error occurred during the scan of the associated object.
func (s *Status) SetErrored(err_msg string) {
	s.Code = StatusErrorCode
	s.CompletedAt = 0
	s.ErroredAt = time.Now().Unix()
	s.State = StatusErrorState
	s.StateMessage = err_msg
}

// SetIgnored() method sets the appropriate fields to indicate that the
// associated object was ignored during the scan, including the timestamp
// at which the object was ignored and an optional state message to
// explain why the object was ignored.
func (s *Status) SetIgnored(state_msg string) {
	s.Code = StatusIgnoredCode
	s.CompletedAt = 0
	s.ErroredAt = 0
	s.IgnoredAt = time.Now().Unix()
	s.State = StatusIgnoredState
	if state_msg != "" {
		s.StateMessage = state_msg
	}
}

// SetRequested() method sets the appropriate fields to indicate that
// a request has been made to an external service to process the data
// from the associated object, including the timestamp at which the
// request was created.
func (s *Status) SetRequested(timestamp int64, state_msg string) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	s.Code = StatusProcessingRequestCode
	s.RequestedAt = timestamp
	s.State = StatusProcessingRequestState
	if state_msg != "" {
		s.StateMessage = state_msg
	}
}

// SetResponded() method sets the appropriate fields to indicate that
// a response was received from an external service for the associated
// object, including the timestamp at which the response was received.
func (s *Status) SetResponded(timestamp int64, state_msg string) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	s.Code = StatusProcessingResponseCode
	s.RespondedAt = timestamp
	s.State = StatusProcessingResponseState
	if state_msg != "" {
		s.StateMessage = state_msg
	}
}

// SetStarted() method sets the appropriate fields to indicate that the
// scan of the associated object has started, including the timestamp at
// which the scan was initiated.
func (s *Status) SetStarted(state_msg string) {
	s.Code = StatusStartCode
	s.StartedAt = time.Now().Unix()
	s.State = StatusStartState
	if state_msg != "" {
		s.StateMessage = state_msg
	}
}

// GetMessageFromResultCode() function returns the message associated
// with the provided result code.
func GetMessageFromResultCode(result_code int) string {
	switch result_code {
	case ResultCleanCode:
		return ResultCleanMsg
	case ResultDirtyCode:
		return ResultDirtyMsg
	case ResultErrorCode:
		return ResultErrorMsg
	default:
		return ResultInitMsg
	}
}
