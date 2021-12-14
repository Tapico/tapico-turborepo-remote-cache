package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	log "github.com/go-kit/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	// Routing and Cloud storage.
	"tapico-turborepo-remote-cache/gcs"

	"github.com/gorilla/mux"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/s3"
)

var logger log.Logger

func GetBucketName(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func getProviderConfig(kind string) (stow.ConfigMap, error) {
	logger.Log("message", "getProviderConfig()")

	var config stow.ConfigMap
	if kind == "s3" {
		logger.Log("message", "getting provider for Amazon S3")
		config = stow.ConfigMap{
			s3.ConfigEndpoint:    "http://127.0.0.1:9000",
			s3.ConfigAccessKeyID: "turborepo",
			s3.ConfigSecretKey:   "turborepo",
			s3.ConfigDisableSSL:  "true",
			s3.ConfigRegion:      "eu-west-1",
		}
	} else if kind == "gcs" {
		logger.Log("message", "getting provider for Google Cloud Storage")
		credFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
		logger.Log(credFile)
		projectID := os.Getenv("GOOGLE_PROJECT_ID")
		logger.Log("google_project_id", projectID)

		config = stow.ConfigMap{
			gcs.ConfigJSON:      credFile,
			gcs.ConfigProjectId: projectID,
		}
	} else {
		logger.Log("message", "getting provider for Local Filesystem")
		configPath, _ := filepath.Abs("./dev/data/filesystem/")
		logger.Log(configPath)

		config = stow.ConfigMap{
			local.ConfigKeyPath: configPath,
		}
	}

	return config, nil
}

func GetContainerByName(name string) (stow.Container, error) {
	logger.Log("message", fmt.Sprintf(`GetContainerByName() name=%s`, name))

	availableCloudProviders := stow.Kinds()
	logger.Log("message", fmt.Sprintf(`GetContainerByName() availableCloudProviders=%s`, availableCloudProviders))

	kind := "gcs"
	config, err := getProviderConfig(kind)
	if err != nil {
		logger.Log("message", "failed to get container config")
		logger.Log(err.Error())
		return nil, err
	}

	// connect
	location, err := stow.Dial(kind, config)
	if err != nil {
		logger.Log("message", "failed to get container instance")
		logger.Log(err.Error())
		return nil, err
	}

	var container stow.Container

	receivedContainer, err := location.Container(name)
	if err != nil {
		logger.Log("message", "failed to fetch existing container with the requested name")
	} else {
		logger.Log("message", "found existing container")
		container = receivedContainer
	}

	if receivedContainer == nil {
		logger.Log("message", "failed to find an existing container")
		createdContainer, err := location.CreateContainer(name)
		if err != nil {
			logger.Log("message", "failed to create container")
			logger.Log(err)
			return nil, err
		}

		logger.Log("message", "create the container for storing cache items")
		container = createdContainer
	}

	logger.Log("message", fmt.Sprintf(`GetContainerByName() id: %s`, container.ID()))
	logger.Log("message", fmt.Sprintf(`GetContainerByName() name: %s`, container.Name()))

	return container, nil
}

func createCacheBlob(name string, teamID string, fileContents io.Reader, fileSize int64) (stow.Item, error) {
	logger.Log("message", "createCacheBlob() called")

	container, err := GetContainerByName(teamID)
	if err != nil {
		return nil, err
	}

	//
	if container == nil {
		logger.Log("message", "failed to lookup container reference")
		return nil, nil
	}

	//
	logger.Log("message", "attempt to save item to cloud storage")
	item, err := container.Put(name, fileContents, fileSize, nil)
	if err != nil {
		logger.Log("message", "failed to save item to cloud storage")
		return nil, err
	}

	logger.Log("message", "attempt to return item")
	itemMetadata, err := item.Metadata()
	if err != nil {
		return nil, err
	}

	for value, name := range itemMetadata {
		logger.Log("name", name, "value", value)
	}

	return item, nil
}

func readCacheBlob(name string, teamID string) (stow.Item, error) {
	logger.Log("message", "readCacheBlob() called")

	container, err := GetContainerByName(teamID)
	if err != nil {
		logger.Log("message", "failed to get container api instance")
		logger.Log(err)
		logger.Log(err.Error())
		return nil, err
	}

	//
	if container == nil {
		logger.Log("message", "failed to lookup container reference")
		logger.Log(err)
		logger.Log(err.Error())
		return nil, nil
	}

	//
	logger.Log("message", "attempt to read item from cloud storage")
	item, err := container.Item(name)
	if err != nil {
		logger.Log("message", "failed to read item from cloud storage")
		if err == stow.ErrNotFound {
			logger.Log("message", "file was not found\n")
		}
		return nil, err
	}

	logger.Log("message", "attempt to return item")
	itemMetadata, err := item.Metadata()
	if err != nil {
		logger.Log(err)
		return nil, err
	}

	for value, name := range itemMetadata {
		logger.Log("name", name, "value", value)
	}

	logger.Log("message", "attempt to return item")
	logger.Log(item.Metadata())

	return item, nil
}

func readCacheItem(w http.ResponseWriter, r *http.Request) {
	logger.Log("message", "readCacheItem()")
	pathParams := mux.Vars(r)

	ctx := r.Context()
	span := oteltrace.SpanFromContext(ctx)
	bag := baggage.FromContext(ctx)

	uk := attribute.Key("username")
	span.AddEvent("handling this...", oteltrace.WithAttributes(uk.String(bag.Member("username").Value())))

	artificateID := ""
	if val, ok := pathParams["artificateID"]; ok {
		artificateID = val
		logger.Log("message", fmt.Sprintf("\nreceived the following artificateID=%s", artificateID))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"error":{"message":"artificateID is missing","code":"required"}}`))
		if err != nil {
			logger.Log("message", err)
		}
		return
	}

	query := r.URL.Query()
	if !query.Has("teamID") {
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"error":{"message":"teamID is missing","code":"required"}}`))
		if err != nil {
			logger.Log("message", err)
		}
		return
	}

	teamID := query.Get("teamID")
	sanitisedteamID := GetBucketName(teamID)
	logger.Log("message", fmt.Sprintf("\nreceived the following teamID=%s sanitisedteamID=%s", teamID, sanitisedteamID))

	// Attempt to return the data from the cloud storage
	item, err := readCacheBlob(artificateID, sanitisedteamID)
	if err != nil {
		logger.Log(err)
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":{"message":"Artifact not found","code":"not_found"}}`))
		return
	}

	fileReference, err := item.Open()
	if err != nil {
		logger.Log(err)
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":{"message":"Artifact not found","code":"not_found"}}`))
		return
	}

	// Attempt to read the file contents of the artificats
	defer fileReference.Close()
	fileBuffer := make([]byte, 4)
	n, err := fileReference.Read(fileBuffer)
	if err != nil {
		logger.Log(err)
		stdlog.Fatal(err)
	}

	logger.Log("message", fmt.Sprintf("\ntotal size of buffer=%d", n))

	w.WriteHeader((http.StatusOK))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Accept, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, PATCH, DELETE")
	w.Write(fileBuffer)
}

func writeCacheItem(w http.ResponseWriter, r *http.Request) {
	logger.Log("message", "writeCacheItem()")
	pathParams := mux.Vars(r)

	ctx := r.Context()
	span := oteltrace.SpanFromContext(ctx)
	bag := baggage.FromContext(ctx)

	uk := attribute.Key("username")
	span.AddEvent("handling this...", oteltrace.WithAttributes(uk.String(bag.Member("username").Value())))

	artificateID := ""
	if val, ok := pathParams["artificateID"]; ok {
		artificateID = val
		logger.Log("message", fmt.Sprintf("\nreceived the following artificateID=%s", artificateID))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":{"message":"artificateID is missing","code":"required"}}`))
		return
	}

	query := r.URL.Query()
	if !query.Has("teamID") {
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":{"message":"teamID is missing","code":"required"}}`))
		return
	}

	teamID := query.Get("teamID")
	sanitisedteamID := GetBucketName(teamID)
	logger.Log("message", "received the following", "teamID", teamID, "sanitisedteamID", sanitisedteamID)

	item, err := createCacheBlob(artificateID, sanitisedteamID, r.Body, r.ContentLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"error":{"message":"failed to save cache item with id %s","code":"internal_error"}}`, artificateID)))
		return
	}

	// Retrieve the url of the uploaded items
	cacheItemURL := item.URL()

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"urls": ["%s/%s"]}`, teamID, cacheItemURL.Path)))
}

func initTracer() *sdktrace.TracerProvider {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	_, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		stdlog.Fatal(err)
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		//sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("tapico-remote-cache-service"))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func main() {
	// Logfmt is a structured, key=val logging format that is easy to read and parse
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	// Direct any attempts to use Go's log package to our structured logger
	stdlog.SetOutput(log.NewStdlibAdapter(logger))
	// Log the timestamp (in UTC) and the callsite (file + line number) of the logging
	// call for debugging in the future.
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "loc", log.DefaultCaller)

	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Log("message", "Error shutting down tracer provider: %v", err)
		}
	}()

	loggingMiddleware := LoggingMiddleware(logger)

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("tapico-remote-cache"))

	// https://api.vercel.com/v8/artifacts/09b4848294e347d8?teamID=team_lMDgmODIeVfSbCQNQPDkX8cF
	api := r.PathPrefix("/v8").Subrouter()
	api.HandleFunc("/artifacts/{artificateID}", readCacheItem).Methods(http.MethodGet)
	api.HandleFunc("/artifacts/{artificateID}", writeCacheItem).Methods(http.MethodPost)
	api.HandleFunc("/artifacts/{artificateID}", writeCacheItem).Methods(http.MethodPut)
	http.Handle("/", r)

	loggedRouter := loggingMiddleware(r)

	// Start server
	address := os.Getenv("LISTEN_ADDRESS")
	if len(address) > 0 {
		err := http.ListenAndServe(address, loggedRouter)
		if err != nil {
			panic(err)
		}
	} else {
		// Default port 8080
		err := http.ListenAndServe("localhost:8080", loggedRouter)
		if err != nil {
			panic(err)
		}
	}
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status int
	// wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func LoggingMiddleware(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Log(
						"err", err,
						"trace", debug.Stack(),
					)
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.Log(
				"status", wrapped.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", time.Since(start),
			)
		}

		return http.HandlerFunc(fn)
	}
}
