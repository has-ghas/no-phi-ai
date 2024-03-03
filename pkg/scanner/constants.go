package scanner

import "time"

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
const KeyCodePending = 1

const KeyStateComplete = "complete"
const KeyStateIgnore = "ignore"
const KeyStateInit = "init"
const KeyStateError = "error"
const KeyStatePending = "pending"

const ScanObjectTypeCommit string = "commit"
const ScanObjectTypeDocument string = "document"
const ScanObjectTypeFile string = "file"
const ScanObjectTypeOrganization string = "organization"
const ScanObjectTypeRepository string = "repository"
const ScanObjectTypeRequestResponse string = "request_and_response"

const ScanRefreshInterval time.Duration = time.Second * 5
