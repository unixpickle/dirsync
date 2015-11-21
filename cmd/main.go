package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/unixpickle/dirsync"
	"github.com/howeyc/gopass"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0],
			"<FTP host> <username> <remote dir> <local dir> <seconds>")
		os.Exit(1)
	}

	ftpHost := os.Args[1]
	username := os.Args[2]
	remoteDir := os.Args[3]
	localDir := os.Args[4]
	seconds, err := strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid interval:", os.Args[5])
		os.Exit(1)
	}

	fmt.Print("Password: ")
	ftpConfig := dirsync.FTPConfig{
		Host: ftpHost,
		User: username,
		Pass: string(gopass.GetPasswd()),
	}
	client := dirsync.NewFTP(ftpConfig)

	syncer := dirsync.Syncer{
		LocalPath: localDir,
		RemotePath: remoteDir,
		Lister: client,
		Downloader: client,
		Interval: time.Duration(seconds) * time.Second,
		Verbose: true,
	}

	syncer.Sync()
}
