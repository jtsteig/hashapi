package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtsteig/hashandstatsservice"
	"github.com/jtsteig/hashmodels"
)

var hashRepo *hashmodels.HashRepository
var service hashandstatsservice.HashStatsService
var shutdownServiceSignal chan bool

func init() {
	filename := "c:\\temp\\testdb.db"
	hashTable := "hashes"
	db, _ := sql.Open("sqlite3", filename)
	hashRepo, initErr := hashmodels.NewHashStore(db, hashTable)
	if initErr != nil {
		log.Fatal(fmt.Sprintf("Error initing the hash repsitory: %q", initErr))
	}
	service = hashandstatsservice.HashStatsService{HashRepository: hashRepo}
	shutdownServiceSignal = make(chan bool, 0)
}

func main() {
	// Setting up the exit for shutdown.
	router := mux.NewRouter().StrictSlash(true)
	router.Path("/hash").HandlerFunc(createHash).Methods("POST")
	router.Path("/hash/{id:[0-9]+}").HandlerFunc(getHash).Methods("GET")
	router.Path("/shutdown").HandlerFunc(shutdownService)
	router.Path("/stats").HandlerFunc(getTotalStats).Methods("GET")

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting server listener: %s", err)
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
		log.Fatalf("Failed to stop server gracefully: %q", err)
	}
	log.Printf("Server shut down.")
}

func getTotalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := service.GetTotalStats()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func shutdownService(w http.ResponseWriter, r *http.Request) {
	shutdownServiceSignal <- true
}

func getHash(w http.ResponseWriter, r *http.Request) {
	countID, convertErr := strconv.Atoi(mux.Vars(r)["id"])
	if convertErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash, getErr := service.GetHash(int64(countID))
	if getErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(hash.HashValue))
}

func createHash(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	countID, saveErr := service.CreateEmptyHashEntry()
	if saveErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.FormatInt(countID, 10)))

	passwordField := r.FormValue("password")
	time.AfterFunc(5*time.Second, func() {
		service.StoreValue(countID, passwordField)
	})
}
