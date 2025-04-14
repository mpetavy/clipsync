package main

import (
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
	BookmarkFile = "bookmarks_3_24_25.html"
)

func init() {
	common.Init("", "", "", "", "", "", "", "", &resources, nil, nil, run, 0)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()

	sb, err := func() (*common.SwapBuffer, error) {
		sb := common.NewSwapBuffer()

		ba, err := os.ReadFile(BookmarkFile)
		if common.Error(err) {
			return nil, err
		}

		_, err = sb.Write(ba)
		if common.Error(err) {
			return nil, err
		}

		w.Header().Set(common.CONTENT_TYPE, common.MimetypeApplicationJson.MimeType)
		w.Header().Set(common.CONTENT_LENGTH, strconv.Itoa(sb.Len()))

		return sb, nil
	}()

	switch err {
	case nil:
		common.Error(common.HTTPResponse(w, r, http.StatusOK, common.MimetypeApplicationJson.MimeType, sb.Len(), sb))
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	common.DebugFunc()
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", getHandler)
	mux.HandleFunc("/set", setHandler)

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
