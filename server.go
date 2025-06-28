package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"github.com/pbnjay/memory"
	"github.com/rs/cors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	httpPort = flag.Int("http.port", 8443, "HTTP server port")
	httpTLS  = flag.Bool("http.tls", true, "HTTP server TLS")
	httpAuth = flag.Bool("http.auth", true, "HTTP authentication enabled")

	server *Server
)

type MuxHandlerFunc func(restURL *common.RestURL, description string, needsAuth bool, handler http.HandlerFunc)

type Server struct {
	MuxHandlerFunc
	Cfg             *ClipsyncCfg
	Mux             *http.ServeMux
	Database        *Database
	Endpoints       *common.StringTable
	EndpointDetails *common.StringTable
	Bookmarks       *Repository[Bookmark]
	Logs            *Repository[Log]
	CrudBookmarks   *CRUD[Bookmark]
	CrudLogs        *CRUD[Log]
}

type ServerHealth struct {
	Product      string                                 `json:"product"`
	Build        string                                 `json:"build"`
	GIT          string                                 `json:"git"`
	PID          int                                    `json:"pid"`
	HealthStatus *orderedmap.OrderedMap[string, string] `json:"healthStatus"`
	Healthy      bool                                   `json:"healthy"`
}

type ServerStatus struct {
	Product       string                                              `json:"product"`
	Build         string                                              `json:"build"`
	GIT           string                                              `json:"git"`
	PID           int                                                 `json:"pid"`
	MemTotal      string                                              `json:"memTotal"`
	MemFree       string                                              `json:"memFree"`
	NumCPU        int                                                 `json:"numCPU"`
	NumGoRoutines int                                                 `json:"numGoRoutines"`
	Flags         *orderedmap.OrderedMap[string, string]              `json:"flags,omitempty"`
	ENV           *orderedmap.OrderedMap[string, string]              `json:"env,omitempty"`
	RestStats     *orderedmap.OrderedMap[string, common.RestURLStats] `json:"restStats"`
}

type Service interface {
	Close() error
	Reset() error
	Health() error
}

func BasicAuth(r *http.Request, username, password string) error {
	common.DebugFunc()

	if !*httpAuth {
		return nil
	}

	err := func() error {
		bm, err := server.Bookmarks.Find(NewWhereTerm().Where(WhereItem{
			Fieldname: BookmarkSchema.Username,
			Operator:  "=",
			Value:     username,
		}))

		if err == ErrNotFound || bm.Password.String() == password {
			return nil
		}

		return ErrBasicAuth
	}()
	if err != nil {
		if common.IsRunningAsExecutable() {
			common.Sleep(time.Second * 5)
		}

		return ErrBasicAuth
	}

	return nil
}

func NewServer() error {
	common.DebugFunc()

	server = &Server{
		Mux: http.NewServeMux(),
	}

	if common.FileExists(*common.FlagCfgFile) {
		cfg, err := common.LoadConfigurationFile[ClipsyncCfg]()
		if common.Error(err) {
			return err
		}

		server.Cfg = cfg
	}

	server.Endpoints = common.NewStringTable()
	server.Endpoints.AddCols("Path", "Method", "Description", "BasicAuth")

	// Database

	//server.HandlerFunc(HeadSync, "Head sync", true, common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.headSyncHandler)))
	//server.HandlerFunc(GetSync, "Get sync", true, common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.getSyncHandler)))
	//server.HandlerFunc(PutSync, "Put sync status", true, common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.putSyncHandler)))
	//server.HandlerFunc(GetStatus, "Get status", true, common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.getStatusHandler)))
	server.HandlerFunc(HeadSync, "Head sync", true, common.ConcurrentLimitHandler(common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.headSyncHandler))))
	server.HandlerFunc(GetSync, "Get sync", true, common.ConcurrentLimitHandler(common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.getSyncHandler))))
	server.HandlerFunc(PutSync, "Put sync status", true, common.ConcurrentLimitHandler(common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.putSyncHandler))))
	server.HandlerFunc(GetStatus, "Get status", true, common.ConcurrentLimitHandler(common.BasicAuthHandler(true, BasicAuth, common.TelemetryHandler(server.getStatusHandler))))

	server.EndpointDetails = common.NewStringTable()
	server.EndpointDetails.AddCols("Path", "Param", "Mandatory", "Description", "Default")

	for _, restURL := range []*common.RestURL{GetSync, PutSync, GetStatus} {
		for _, param := range restURL.Params {
			server.EndpointDetails.AddCols(restURL.Endpoint, param.Name, strconv.FormatBool(param.Mandatory), param.Description, param.Default)
		}
	}

	// database

	var err error

	server.Database, err = NewDatabase()
	if common.Error(err) {
		return err
	}

	server.Bookmarks, err = NewRepository[Bookmark](server.Database)
	if common.Error(err) {
		return err
	}

	server.CrudBookmarks, err = NewCrud[Bookmark](server.HandlerFunc, server.Bookmarks, BasicAuth, REST_BOOKMARKS)
	if common.Error(err) {
		return err
	}

	server.Logs, err = NewRepository[Log](server.Database)
	if common.Error(err) {
		return err
	}

	server.CrudLogs, err = NewCrud[Log](server.HandlerFunc, server.Logs, BasicAuth, REST_LOGS)
	if common.Error(err) {
		return err
	}

	// Server healthy?

	err = func() error {
		serverHealth, err := server.Health()
		if common.Error(err) {
			return err
		}

		if !serverHealth.Healthy {
			return fmt.Errorf("%s not healthy", strings.ToUpper(common.Title()))
		}

		return nil
	}()

	if common.Error(err) {
		return err
	}

	common.Debug("Endpoints\n" + server.Endpoints.Markdown())
	common.Debug("Endpoint details\n" + server.EndpointDetails.Markdown())

	return nil
}

func (server *Server) Start() error {
	var tlsConfig *tls.Config
	var err error

	if *httpTLS {
		tlsConfig, err = common.NewTlsConfigFromFlags()
		if common.Error(err) {
			return err
		}
	}

	err = common.HTTPServerStart(*httpPort, tlsConfig, cors.Default().Handler(server.Mux))
	if common.Error(err) {
		return err
	}

	return nil
}

func (server *Server) Stop() {
	common.DebugFunc()

	common.Error(server.Database.Close())
	common.Error(common.HTTPServerStop())
}

func (server *Server) HandlerFunc(restURL *common.RestURL, description string, needsAuth bool, handler http.HandlerFunc) {
	common.DebugFunc("%v %v %v", restURL.MuxString(), description, needsAuth)

	server.Endpoints.AddCols(restURL.Endpoint, restURL.Method, description, fmt.Sprintf("%v", needsAuth))

	server.Mux.HandleFunc(restURL.MuxString(), handler)
}

func (server *Server) status() (*ServerStatus, error) {
	serverStatus := &ServerStatus{
		Product:       common.TitleVersion(true, true, true),
		Build:         common.App().Build,
		GIT:           common.App().Git,
		PID:           os.Getpid(),
		MemTotal:      common.FormatMemory(memory.TotalMemory()),
		MemFree:       common.FormatMemory(memory.FreeMemory()),
		NumCPU:        runtime.NumCPU(),
		NumGoRoutines: runtime.NumGoroutine(),
		Flags:         orderedmap.New[string, string](),
		ENV:           orderedmap.New[string, string](),
		RestStats:     orderedmap.New[string, common.RestURLStats](),
	}

	// Flags

	flag.VisitAll(func(f *flag.Flag) {
		serverStatus.Flags.Set(f.Name, common.HideSecretFlags(f.Name, f.Value.String()))
	})

	// ENV

	envs := []string{}
	for _, env := range os.Environ() {
		envs = append(envs, env)
	}

	sort.Strings(envs)

	for _, env := range envs {
		splits := common.Split(env, "=")
		serverStatus.ENV.Set(splits[0], common.HideSecretFlags(splits[0], splits[1]))
	}

	// RestStats

	for _, restURL := range []*common.RestURL{GetSync, PutSync, GetStatus} {
		serverStatus.RestStats.Set(strings.ReplaceAll(restURL.MuxString(), " ", ":"), restURL.Statistics())
	}

	return serverStatus, nil
}

func (server *Server) Health() (*ServerHealth, error) {
	serverHealth := &ServerHealth{
		Product:      common.TitleVersion(true, true, true),
		Build:        common.App().Build,
		GIT:          common.App().Git,
		PID:          os.Getpid(),
		HealthStatus: orderedmap.New[string, string](),
		Healthy:      false,
	}

	serverHealth.Healthy = true

	syncHealth := common.NewSyncOf(serverHealth.HealthStatus)
	wg := sync.WaitGroup{}

	for _, service := range []Service{server.Database} {
		wg.Add(1)

		go func(service Service) {
			start := time.Now()

			defer common.UnregisterGoRoutine(common.RegisterGoRoutine(1))

			defer func() {
				wg.Done()
			}()

			err := service.Health()

			common.Error(syncHealth.RunSynchronized(func(healthStatus *orderedmap.OrderedMap[string, string]) error {
				name := reflect.TypeOf(service).Elem().Name()
				if err != nil {
					healthStatus.Set(name, err.Error())
					if serverHealth.Healthy {
						serverHealth.Healthy = false
					}
				} else {
					healthStatus.Set(name, fmt.Sprintf("ok, took %v", time.Since(start)))
				}

				return nil
			}))
		}(service)
	}

	wg.Wait()

	return serverHealth, nil
}

func (server *Server) Verify() error {
	return nil
}
