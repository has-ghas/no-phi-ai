package scannerv2

const DelimitDocumentID string = "__"

const IgnoreReasonDefault string = "ignored_by_default"
const IgnoreReasonDirPath string = "directory_path"
const IgnoreReasonFileExtensionIgnoredByConfig string = "file_extension_ignored_by_config"
const IgnoreReasonFileExtensionIgnoredByPolicy string = "file_extension_ignored_by_policy"
const IgnoreReasonFileExtensionNotIncluded string = "file_extension_not_included"
const IgnoreReasonFileIsBinary string = "file_is_binary"
const IgnoreReasonFileIsEmpty string = "file_is_empty"
const IgnoreReasonFileName string = "file_name"
const IgnoreReasonFilePath string = "file_path"

const KeyCodeComplete = 2
const KeyCodeIgnore = 0
const KeyCodeInit = -2
const KeyCodeError = -1
const KeyCodeWarning = 1

const KeyStateComplete = "complete"
const KeyStateIgnore = "ignore"
const KeyStateInit = "init"
const KeyStateError = "error"
const KeyStateWarning = "warning"

const ResultCleanCode int = 200
const ResultCleanMsg string = "clean"
const ResultErrorCode int = 500
const ResultErrorMsg string = "error"
const ResultInitCode int = 0
const ResultInitMsg string = "unknown"
const ResultDirtyCode int = 400
const ResultDirtyMsg string = "dirty"

const ScanMetricsRefreshSeconds int = 10
const ScanObjectTypeCommit string = "commit"
const ScanObjectTypeDocument string = "document"
const ScanObjectTypeFile string = "file"
const ScanObjectTypeOrganization string = "organization"
const ScanObjectTypeRepository string = "repository"

const StatusCompleteCode int = 100
const StatusCompleteState string = "complete"
const StatusErrorCode int = -2
const StatusErrorState string = "error"
const StatusIgnoredCode int = -1
const StatusIgnoredState string = "ignored"
const StatusInitCode int = 0
const StatusInitState string = "init"
const StatusProcessingRequestCode int = 50
const StatusProcessingRequestState string = "processing_request"
const StatusProcessingResponseCode int = 75
const StatusProcessingResponseState string = "processing_response"
const StatusStartCode int = 25
const StatusStartState string = "start"
