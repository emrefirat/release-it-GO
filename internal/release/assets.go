package release

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// ResolveAssets resolves glob patterns into a list of file paths.
func ResolveAssets(patterns []string) ([]string, error) {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		files = append(files, matches...)
	}
	return files, nil
}

// knownContentTypes maps file extensions to MIME types for common release assets.
var knownContentTypes = map[string]string{
	".zip":    "application/zip",
	".tar.gz": "application/gzip",
	".tgz":    "application/gzip",
	".gz":     "application/gzip",
	".bz2":    "application/x-bzip2",
	".xz":     "application/x-xz",
	".dmg":    "application/x-apple-diskimage",
	".deb":    "application/vnd.debian.binary-package",
	".rpm":    "application/x-rpm",
	".exe":    "application/vnd.microsoft.portable-executable",
	".msi":    "application/x-msi",
	".apk":    "application/vnd.android.package-archive",
	".sha256": "text/plain",
	".sig":    "application/pgp-signature",
	".asc":    "application/pgp-signature",
}

// DetectContentType determines the MIME type of a file based on its extension.
func DetectContentType(filePath string) string {
	name := strings.ToLower(filePath)

	// Check compound extensions first (e.g., .tar.gz)
	for ext, ct := range knownContentTypes {
		if strings.HasSuffix(name, ext) {
			return ct
		}
	}

	// Try standard mime type detection
	ext := filepath.Ext(name)
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}

	return "application/octet-stream"
}
