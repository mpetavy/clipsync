package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mpetavy/common"
	"net/http"
	"strings"
	"time"
)

const (
	API_ENDPOINT   = "/api/v1"
	REST_SYNC      = API_ENDPOINT + "/sync"
	REST_BOOKMARKS = API_ENDPOINT + "/bookmarks"
	REST_LOGS      = API_ENDPOINT + "/logs"
	REST_STATUS    = API_ENDPOINT + "/status"
)

var (
	ErrBasicAuth = fmt.Errorf("BasicAuth failed")

	HeadSync  = common.NewRestURL(http.MethodHead, REST_SYNC)
	GetSync   = common.NewRestURL(http.MethodGet, REST_SYNC)
	PutSync   = common.NewRestURL(http.MethodPut, REST_SYNC)
	GetStatus = common.NewRestURL(http.MethodGet, REST_STATUS)
)

func (server *Server) headSyncHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	common.Error(common.HTTPResponse(w, r, http.StatusOK, "", 0, nil))
}

func (server *Server) getSyncHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	var err error
	bm, err := func() (*Bookmark, error) {
		bm, err := server.CrudSync.Repository.Find(NewWhereTerm().Where(WhereItem{BookmarkSchema.Email, "=", "dummy"}))
		if common.Error(err) {
			return nil, err
		}

		return bm, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, len(bm.Payload.String()), strings.NewReader(bm.Payload.String())))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (server *Server) putSyncHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	var err error
	//err := func() error {
	//	ba, err := common.ReadBody(r.Body)
	//	if common.Error(err) {
	//		return err
	//	}
	//
	//	bm, err := NewBookmark()
	//	if common.Error(err) {
	//		return err
	//	}
	//
	//	bm.Email.SetString("dummy")
	//	bm.Payload.SetString(string(ba))
	//
	//	err = server.CrudSync.Repository.Save([]Bookmark{*bm})
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
