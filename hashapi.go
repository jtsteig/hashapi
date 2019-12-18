package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ascarter/requestid"
	"github.com/gorilla/mux"
)

var hashRepo *HashRepository
var service HashStatsService
var shutdownServiceSignal chan bool

func init() {
	PrintInfoLog("Startup", "Initializing service.")

	filename := "c:\\temp\\testdb.db"
	hashTable := "hashes"
	db, _ := sql.Open("sqlite3", filename)
	hashRepo, initErr := NewHashRepository(db, hashTable)
	if initErr != nil {
		PrintFatalLog("Startup", fmt.Sprintf("Error initing the hash repsitory: %q", initErr))
	}
	service = HashStatsService{HashRepository: hashRepo}
	shutdownServiceSignal = make(chan bool, 0)
}

func addRoute(router *mux.Router, path string, handlerFn func(w http.ResponseWriter, r *http.Request), method string) {
	handler := http.HandlerFunc(handlerFn)
	router.Handle(path, requestid.RequestIDHandler(handler)).Methods(method)
}

func main() {
	// Setting up the exit for shutdown.
	router := mux.NewRouter().StrictSlash(true)
	addRoute(router, "/hash", createHash, "POST")
	addRoute(router, "/hash/{id:[0-9]+}", getHash, "GET")
	addRoute(router, "/shutdown", shutdownService, "GET")
	addRoute(router, "/stats", getTotalStats, "GET")

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			PrintFatalLog("Startup", fmt.Sprintf("Error starting server listener: %s", err))
		}
	}()

	<-shutdownServiceSignal

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	// Clean up and handle cancelling.
	defer func() {
		service.Close()
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		PrintFatalLog("Shutdown", fmt.Sprintf("Failed to stop server gracefully: %q", err))
	}
	PrintInfoLog("Shutdown", "Server shut down.")
}

func getTotalStats(w http.ResponseWriter, r *http.Request) {
	rid, getRequestIDSuccess := requestid.FromContext(r.Context())
	if !getRequestIDSuccess {
		PrintInfoLog(rid, "Error setting requestID for getTotalStats")
	}

	PrintInfoLog(rid, "Getting total stats.")
	stats, err := service.GetTotalStats()
	if err != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failure getting total stats: %q", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if serErr := json.NewEncoder(w).Encode(stats); serErr != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failed to serialize TotalStats: %q", serErr))
	}
}

func shutdownService(w http.ResponseWriter, r *http.Request) {
	rid, getRequestIDSuccess := requestid.FromContext(r.Context())
	if !getRequestIDSuccess {
		PrintInfoLog(rid, "Error setting requestID for shutdownService")
	}
	PrintInfoLog(rid, "Request recieved to shutdown service.")
	shutdownServiceSignal <- true
}

func getHash(w http.ResponseWriter, r *http.Request) {
	rid, getRequestIDSuccess := requestid.FromContext(r.Context())
	if getRequestIDSuccess {
		PrintInfoLog(rid, "Error setting requestID for getHash")
	}

	PrintInfoLog(rid, "getHash request recieved.")
	idVal := mux.Vars(r)["id"]
	countID, convertErr := strconv.Atoi(idVal)
	if convertErr != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failed to convert id (%q) to a proper integer: %q", idVal, convertErr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash, getErr := service.GetHash(int64(countID))
	if getErr != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failed to get hash for countID (%q): %q", countID, getErr))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	PrintInfoLog(rid, fmt.Sprintf("Returning value for getHash: (%q)", hash.HashValue))
	w.Write([]byte(hash.HashValue))
}

func createHash(w http.ResponseWriter, r *http.Request) {
	rid, getRequestIDSuccess := requestid.FromContext(r.Context())
	if !getRequestIDSuccess {
		PrintInfoLog(rid, "Error setting requestID for createHash")
	}
	PrintInfoLog(rid, "createHash request recieved.")
	if err := r.ParseForm(); err != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failed to parse formdata for createHash request: %q", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	countID, saveErr := service.CreateEmptyHashEntry()
	if saveErr != nil {
		PrintErrorLog(rid, fmt.Sprintf("Failed to create row for CreateEmptyHashEntry with countID(%q): %q", countID, saveErr))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.FormatInt(countID, 10)))

	passwordField := r.FormValue("password")

	PrintInfoLog(rid, fmt.Sprintf("Original passwordfield: %q", passwordField))
	time.AfterFunc(5*time.Second, func() {
		PrintInfoLog(rid, fmt.Sprintf("Storing hash for password %q", passwordField))
		service.StoreValue(countID, passwordField)
	})
}
