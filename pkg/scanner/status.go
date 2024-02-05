package scanner

import "time"

// Status struct is used to track the status of a scan for the
// associated ScanObject, where Status is embedded in ScanObject.
type Status struct {
	Code          int    `json:"code"`
	CompletedAt   int64  `json:"completed_at"`
	ErroredAt     int64  `json:"errored_at"`
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
		RequestedAt:   0,
		RespondedAt:   0,
		ResultCode:    ResultCleanCode,
		ResultMessage: ResultInitMsg,
		StartedAt:     0,
		State:         StatusInitState,
		StateMessage:  "",
	}
}

// SetCompleted() method sets the appropriate fields to indicate that
// the code, message (optional), state, and timestamp for the completed
// scan of the associated object.
func (s *Status) SetCompleted(result_code int, msg string) {
	s.Code = StatusCompleteCode
	s.CompletedAt = time.Now().Unix()
	s.ErroredAt = 0
	s.ResultCode = result_code
	s.ResultMessage = GetMessageFromResultCode(result_code)
	s.State = StatusCompleteState
	if msg != "" {
		s.StateMessage = msg
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

// SetRequested() method sets the appropriate fields to indicate that
// a request has been made to an external service to process the data
// from the associated object, including the timestamp at which the
// request was created.
func (s *Status) SetRequested(timestamp int64, msg string) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	s.Code = StatusProcessingRequestCode
	s.RequestedAt = timestamp
	s.State = StatusProcessingRequestState
	if msg != "" {
		s.StateMessage = msg
	}
}

// SetResponded() method sets the appropriate fields to indicate that
// a response was received from an external service for the associated
// object, including the timestamp at which the response was received.
func (s *Status) SetResponded(timestamp int64, msg string) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	s.Code = StatusProcessingResponseCode
	s.RespondedAt = timestamp
	s.State = StatusProcessingResponseState
	if msg != "" {
		s.StateMessage = msg
	}
}

// SetStarted() method sets the appropriate fields to indicate that the
// scan of the associated object has started, including the timestamp at
// which the scan was initiated.
func (s *Status) SetStarted(msg string) {
	s.Code = StatusStartCode
	s.StartedAt = time.Now().Unix()
	s.State = StatusStartState
	if msg != "" {
		s.StateMessage = msg
	}
}

// GetMessageFromResultCode() function returns the message associated
// with the provided result code.
func GetMessageFromResultCode(code int) string {
	switch code {
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
