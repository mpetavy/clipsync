package main

import (
	"embed"
	"github.com/mpetavy/common"
	"github.com/rs/cors"
	"io"
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

func getBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	file, size, err := func() (io.ReadCloser, int64, error) {
		if !common.FileExists(BookmarkFile) {
			return nil, 0, &common.ErrFileNotFound{
				FileName: BookmarkFile,
			}
		}

		size, err := common.FileSize(BookmarkFile)
		if common.Error(err) {
			return nil, 0, err
		}

		file, err := os.Open(BookmarkFile)
		if common.Error(err) {
			return nil, 0, err
		}

		w.Header().Set(common.CONTENT_TYPE, common.MimetypeApplicationJson.MimeType)
		w.Header().Set(common.CONTENT_LENGTH, strconv.Itoa(int(size)))

		return file, size, nil
	}()

	switch err {
	case nil:
		defer func() {
			common.Error(file.Close())
		}()

		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, int(size), file))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func setBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	var err error
	//err := func() error {
	//	file, err := os.OpenFile(BookmarkFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, common.DefaultFileMode)
	//	if common.Error(err) {
	//		return err
	//	}
	//
	//	defer func() {
	//		common.Error(file.Close())
	//	}()
	//
	//	_, err = io.Copy(file, r.Body)
	//	defer func() {
	//		common.Error(r.Body.Close())
	//	}()
	//
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
	mux.HandleFunc("/get", getBookmarkHandler)
	mux.HandleFunc("/set", setBookmarkHandler)

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
