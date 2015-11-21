package dirsync

type FileInfo struct {
	IsDir    bool
	Path     string
	FileSize uint64
}

type Lister interface {
	List(path string) ([]FileInfo, error)
}

type Downloader interface {
	Download(remotePath, localPath string) error
}
