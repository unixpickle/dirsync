package dirsync

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// A Syncer keeps a local folder up to date with a remote server.
type Syncer struct {
	LocalPath  string
	RemotePath string

	Lister     Lister
	Downloader Downloader

	Interval time.Duration

	Verbose bool
}

// Sync runs an infinite sync loop.
// The loop reads remote listings, downloads any locally missing files, and deletes any
// extraneous local files.
// If a synchronization fails, this will return the error.
func (s Syncer) Sync() error {
	for {
		nextTimeout := time.After(s.Interval)
		if err := s.syncOnce(); err != nil {
			if s.Verbose {
				log.Println("Sync failed:", err)
			}
			return err
		}
		<-nextTimeout
	}
}

// syncOnce performs one sync iteration.
// This returns the first error that occurs.
func (s Syncer) syncOnce() error {
	a, err := s.computeAgenda()
	if err != nil {
		return err
	}

	for _, path := range a.delete {
		if s.Verbose {
			log.Println("Removing local file:", path)
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	for i, remote := range a.download {
		if err := s.download(remote, a.downloadDest[i]); err != nil {
			return err
		}
	}

	return nil
}

// download recursively copies a remote directory to the local path.
func (s Syncer) download(remote FileInfo, local string) error {
	if s.Verbose {
		log.Println("Downloading:", remote, "->", local)
	}

	if !remote.IsDir {
		return s.Downloader.Download(remote.Path, local)
	}

	listing, err := s.Lister.List(remote.Path)
	if err != nil {
		return err
	}

	for _, entry := range listing {
		dest := filepath.Join(local, entry.Name())
		if err := s.download(entry, dest); err != nil {
			return err
		}
	}
	return nil
}

// agenda is a list of things which need to be deleted and downloaded.
type agenda struct {
	delete       []string
	download     []FileInfo
	downloadDest []string
}

type agendaSearchNode struct {
	localPath  string
	remotePath string
}

// computeAgenda generates an agenda to correct the differences between
// the local and remote directory.
func (s Syncer) computeAgenda() (*agenda, error) {
	agenda := &agenda{[]string{}, []FileInfo{}, []string{}}

	// Perform a breadth-first search over the remote directory.
	nodes := []agendaSearchNode{{s.LocalPath, s.RemotePath}}
	for len(nodes) > 0 {
		n := nodes[0]
		nodes = nodes[1:]

		localListing, err := ioutil.ReadDir(n.localPath)
		if err != nil {
			return nil, err
		}
		remoteListing, err := s.Lister.List(n.remotePath)
		if err != nil {
			return nil, err
		}

		for _, local := range localListing {
			var remoteMatch *FileInfo
			for _, remote := range remoteListing {
				if localRemoteFilesEqual(local, remote) {
					remoteMatch = &remote
					break
				}
			}
			localPath := filepath.Join(n.localPath, local.Name())
			if remoteMatch == nil {
				agenda.delete = append(agenda.delete, localPath)
			} else if local.IsDir() {
				node := agendaSearchNode{
					localPath:  localPath,
					remotePath: remoteMatch.Path,
				}
				nodes = append(nodes, node)
			}
		}

		for _, remote := range remoteListing {
			foundMatch := false
			for _, local := range localListing {
				if localRemoteFilesEqual(local, remote) {
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				agenda.download = append(agenda.download, remote)
				dest := filepath.Join(n.localPath, remote.Name())
				agenda.downloadDest = append(agenda.downloadDest, dest)
			}
		}
	}

	return agenda, nil
}

func localRemoteFilesEqual(local os.FileInfo, remote FileInfo) bool {
	if local.Name() != remote.Name() {
		return false
	}
	if local.IsDir() != remote.IsDir {
		return false
	}
	if !local.IsDir() && uint64(local.Size()) != remote.FileSize {
		return false
	}
	return true
}
