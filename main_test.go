package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestMain(m *testing.M) {
	common.RunTests(m)
}

func address() string {
	var protocol string

	if *httpTLS {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://localhost:%d", protocol, *httpPort)
}

func TestBookmarks(t *testing.T) {
	r, ba, err := common.HTTPRequest(nil, common.MillisecondToDuration(*common.FlagHTTPTimeout), GetBookmarks.Method, GetBookmarks.URL(address()), nil, nil, *httpUsername, *httpPassword, nil, http.StatusOK)

	fmt.Printf("%s\n", ba)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.StatusCode)
}
