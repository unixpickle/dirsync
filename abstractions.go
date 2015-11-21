package dirsync

import "path"

type FileInfo struct {
	// IsDir is true if and only if the remote file is a directory.
	IsDir bool

	// Path is a slash-separated path in the remote filesystem.
	Path string

	// FileSize is the size of the file (in bytes).
	// This field is not used for directories.
	FileSize uint64
}

// Name returns the basename of the path.
func (f FileInfo) Name() string {
	return path.Base(f.Path)
}

// A Lister provides remote directory listings.
type Lister interface {
	// List reads the contents of a remote (slash-separated) path.
	// If an error occurs, this should return nil, err.
	List(path string) ([]FileInfo, error)
}

// A Downloader downloads remote files (but not directories).
type Downloader interface {
	// Download the file located at a remote (slash-separated) path.
	// The file should be saved at the given local path.
	Download(remotePath, localPath string) error
}
