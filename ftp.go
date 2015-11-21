package dirsync

import (
	"io"
	"os"
	"path"

	"github.com/jlaffaye/ftp"
)

type FTPConfig struct {
	Host string
	User string
	Pass string
}

// FTP is a Lister and Downloader which operates on a remote host.
type FTP struct {
	config     FTPConfig
	connection *ftp.ServerConn
}

// NewFTP creates a new FTP instance using a given configuration.
// It will not automatically connect to the remote end.
func NewFTP(c FTPConfig) *FTP {
	return &FTP{c, nil}
}

// Disconnect shuts down the FTP connection.
func (f *FTP) Disconnect() {
	if f.connection != nil {
		f.connection.Quit()
		f.connection = nil
	}
}

// List returns the contents of a directory.
// If the connection fails, this returns an error.
func (f *FTP) List(dir string) ([]FileInfo, error) {
	if err := f.ensureConnected(); err != nil {
		return nil, err
	}

	entries, err := f.connection.List(dir)
	if err != nil {
		f.Disconnect()
		return nil, err
	}

	res := make([]FileInfo, len(entries))
	for i, entry := range entries {
		res[i] = FileInfo{
			Path:     path.Join(dir, entry.Name),
			IsDir:    entry.Type == ftp.EntryTypeFolder,
			FileSize: entry.Size,
		}
	}

	return res, nil
}

// Download gets a remote file and stores it at a local path.
// If the connection fails or the local file cannot be written, this returns an error.
func (f *FTP) Download(remotePath, localPath string) error {
	if err := f.ensureConnected(); err != nil {
		return err
	}

	reader, err := f.connection.Retr(remotePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(localPath)
		return err
	}

	return nil
}

func (f *FTP) ensureConnected() error {
	if f.connection != nil {
		if f.connection.NoOp() != nil {
			f.Disconnect()
		}
	}
	if f.connection == nil {
		return f.connect()
	}
	return nil
}

func (f *FTP) connect() error {
	var err error
	if f.connection, err = ftp.Connect(f.config.Host); err != nil {
		return err
	}

	if err = f.connection.Login(f.config.User, f.config.Pass); err != nil {
		f.Disconnect()
		return err
	}

	return nil
}
