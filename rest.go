package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mpetavy/common"
	"github.com/spyzhov/ajson"
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
		err := GetSync.Validate(r)
		if common.Error(err) {
			return nil, err
		}

		username, _, _ := r.BasicAuth()

		bm, err := server.CrudBookmarks.Repository.Find(NewWhereTerm().Where(WhereItem{BookmarkSchema.Username, "=", username}))
		if common.Error(err) {
			return nil, err
		}

		return bm, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, len(bm.Payload.String()), strings.NewReader(bm.Payload.String())))
	default:
		http.Error(w, err.Error(), http.StatusNoContent)
	}
}

func (server *Server) putSyncHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	err := server.Database.RunSynchronized(func() error {
		err := PutSync.Validate(r)
		if common.Error(err) {
			return err
		}

		username, password, _ := r.BasicAuth()

		ba, err := common.ReadBody(r.Body)
		if common.Error(err) {
			return err
		}

		_, err = ajson.JSONPath(ba, "$..dateAdded")
		if err != nil {
			return err
		}

		//if len(jo) == 3 {
		//	common.Info("Do not save empty bookmark list")
		//
		//	return nil
		//}

		bm, err := server.CrudBookmarks.Repository.Find(NewWhereTerm().Where(WhereItem{BookmarkSchema.Username, "=", username}))
		switch err {
		case nil:
		case ErrNotFound:
			bm, err = NewBookmark()
			if common.Error(err) {
				return err
			}
		default:
			if common.Error(err) {
				return err
			}

		}

		bm.Username.SetString(username)
		bm.Password.SetString(password)
		bm.Payload.SetString(string(ba))

		err = server.CrudBookmarks.Repository.Save(bm)
		if common.Error(err) {
			return err
		}

		return nil
	})

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
