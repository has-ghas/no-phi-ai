package scanner

const DelimitDocumentID string = "__"

const ResultCleanCode int = 200
const ResultCleanMsg string = "clean"
const ResultErrorCode int = 500
const ResultErrorMsg string = "error"
const ResultInitCode int = 0
const ResultInitMsg string = "unknown"
const ResultDirtyCode int = 400
const ResultDirtyMsg string = "dirty"

const ScanObjectTypeCommit string = "commit"
const ScanObjectTypeFile string = "file"
const ScanObjectTypeOrganization string = "organization"
const ScanObjectTypeRepository string = "repository"

const StatusCompleteCode int = 100
const StatusCompleteState string = "complete"
const StatusErrorCode int = -1
const StatusErrorState string = "error"
const StatusInitCode int = 0
const StatusInitState string = "init"
const StatusProcessingRequestCode int = 50
const StatusProcessingRequestState string = "processing_request"
const StatusProcessingResponseCode int = 75
const StatusProcessingResponseState string = "processing_response"
const StatusStartCode int = 25
const StatusStartState string = "start"
