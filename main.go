package main

import (
	"embed"
	"github.com/mpetavy/common"
)

//go:embed go.mod
var resources embed.FS

const (
	//BookmarkFile = "bookmarks_3_24_25.html"
	BookmarkFile = "testdata/bookmarks.json"
)

func init() {
	common.Init("", "", "", "", "Syncs bookmarks", "", "", "", &resources, start, stop, nil, 0)
}

func start() error {
	common.DebugFunc()

	err := common.IsPortAvailable("tcp", *httpPort)
	if common.Error(err) {
		return err
	}

	err = NewServer()
	if common.Error(err) {
		return err
	}

	err = server.Verify()
	if common.Error(err) {
		return err
	}

	err = server.Start()
	if common.Error(err) {
		return err
	}

	return nil
}

func stop() error {
	common.DebugFunc()

	server.Stop()

	return nil
}

func main() {
	common.Run(common.MandatoryFlags("db.file"))
}
