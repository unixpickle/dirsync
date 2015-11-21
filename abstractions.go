package dirsync

type FileInfo struct {
	IsDir    bool
	Path     string
	FileSize int64
}

type Lister interface {
	List(path string) []FileInfo
}

type Downloader interface {
	Download(remotePath, localPath string) bool
}
