package main

import (
	"bytes"
	"embed"
	"github.com/mpetavy/common"
	"github.com/rs/cors"
	"net/http"
	"os"
	"strconv"
)

//go:embed go.mod
var resources embed.FS

const (
	//BookmarkFile = "bookmarks_3_24_25.html"
	BookmarkFile = "bookmarks.json"
)

func init() {
	common.Init("", "", "", "", "", "", "", "", &resources, nil, nil, run, 0)
}

func restoreBookmarks(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	ba, err := func() ([]byte, error) {
		if !common.FileExists(BookmarkFile) {
			return nil, &common.ErrFileNotFound{
				FileName: BookmarkFile,
			}
		}

		ba, err := os.ReadFile(BookmarkFile)
		if common.Error(err) {
			return nil, err
		}

		w.Header().Set(common.CONTENT_TYPE, common.MimetypeApplicationJson.MimeType)
		w.Header().Set(common.CONTENT_LENGTH, strconv.Itoa(len(ba)))

		return ba, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, len(ba), bytes.NewReader(ba)))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func backupBookmarks(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	var err error
	//err := func() error {
	//	ba, err := common.ReadBody(r.Body)
	//	if common.Error(err) {
	//		return err
	//	}
	//
	//	err = os.WriteFile(BookmarkFile, ba, common.DefaultFileMode)
	//	if common.Error(err) {
	//		return err
	//	}
	//
	//	return nil
	//}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, "", 0, nil))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/restoreBookmarks", restoreBookmarks)
	mux.HandleFunc("/backupBookmarks", backupBookmarks)

	handler := cors.Default().Handler(mux)

	err := http.ListenAndServe(":8080", handler)
	if common.Error(err) {
		return err
	}

	return nil
}

func main() {
	common.Run(nil)
}
