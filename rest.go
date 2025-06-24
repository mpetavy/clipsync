package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mpetavy/common"
	"net/http"
	"os"
	"time"
)

const (
	API_ENDPOINT   = "/api/v1"
	REST_BOOKMARKS = API_ENDPOINT + "/bookmarks"
	REST_STATUS    = API_ENDPOINT + "/status"
)

var (
	ErrBasicAuth = fmt.Errorf("BasicAuth failed")

	GetBookmarks = common.NewRestURL(http.MethodGet, REST_BOOKMARKS)
	PutBookmarks = common.NewRestURL(http.MethodPut, REST_BOOKMARKS)
	GetStatus    = common.NewRestURL(http.MethodGet, REST_STATUS)
)

func (server *Server) getBookmarksHandler(w http.ResponseWriter, r *http.Request) {
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

		return ba, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, len(ba), bytes.NewReader(ba)))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (server *Server) putBookmarksHandler(w http.ResponseWriter, r *http.Request) {
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

func (server *Server) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	defer func(start time.Time) {
		GetStatus.UpdateStats(start)
	}(time.Now())

	ba, err := func() ([]byte, error) {
		err := GetStatus.Validate(r)
		if common.Error(err) {
			return nil, err
		}

		serverStatus, err := server.status()
		if common.Error(err) {
			return nil, err
		}

		ba, err := json.MarshalIndent(serverStatus, "", "    ")
		if common.Error(err) {
			return nil, err
		}

		return ba, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, len(ba), bytes.NewReader(ba)))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
