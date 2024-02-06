package scanner

import "path/filepath"

// IgnoreFilenames is a map of file names that should not be scanned.
var IgnoreFilenames = map[string]bool{
	"":           true,
	".git":       true,
	".gitignore": true,
	".gitkeep":   true,
}

// IgnoreExtensions is a map of file extensions that should not be scanned.
var IgnoreExtensions = map[string]bool{
	".":    true,
	".jpg": true,
}

// IgnoreFilePath() function checks if a file path should be ignored
// (i.e. not scanned) based on the file extansion and name. Returns
// ignore = true if the file should be ignored, and ignore = false
// if the file should be scanned. Also returns a reason string that
// explains the justification for ignoring the file, if applicable.
func IgnoreFilePath(path string) (ignore bool, reason string) {
	// explicitly set defaults for return values
	ignore = false
	reason = ""

	// get the file name from the path
	file_name := filepath.Base(path)
	// check if the filename is in the IgnoreFilenames map
	if ignore_file_name, exists := IgnoreFilenames[file_name]; exists && ignore_file_name {
		// ignore the file / do not scan
		ignore = true
		reason = IgnoreReasonFileName
		return
	}

	// get the file extension from the path
	file_extension := filepath.Ext(path)
	if ignore_file_extension, exists := IgnoreExtensions[file_extension]; exists && ignore_file_extension {
		// ignore the file / do not scan
		ignore = true
		reason = IgnoreReasonFileExtension
		return
	}

	return
}
