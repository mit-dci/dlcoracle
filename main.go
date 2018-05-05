package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/logging"
	"github.com/gertjaap/dlcoracle/publisher"
	"github.com/gertjaap/dlcoracle/routes"
	"github.com/gertjaap/dlcoracle/store"

	"github.com/awnumar/memguard"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	logging.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	logging.Info.Println("MIT-DCI Discreet Log Oracle starting...")

	key, err := crypto.ReadKeyFile("data/privkey.hex")
	if err != nil {
		logging.Error.Fatal("Could not open or create keyfile:", err)
		os.Exit(1)
	}
	crypto.StoreKeys(key)
	// Tell memguard to listen out for interrupts, and cleanup in case of one.
	memguard.CatchInterrupt(func() {
		fmt.Println("Interrupt signal received. Exiting...")
	})
	// Make sure to destroy all LockedBuffers when returning.
	defer memguard.DestroyAll()

	store.Init()
	logging.Info.Println("Initialized store...")

	publisher.Init()
	logging.Info.Println("Started publisher...")

	r := mux.NewRouter()
	r.HandleFunc("/api/datasources", routes.ListDataSourcesHandler)
	r.HandleFunc("/api/datasource/{id}/value", routes.DataSourceValueHandler)
	r.HandleFunc("/api/pubkey", routes.PubKeyHandler)
	r.HandleFunc("/api/rpoint/{datasource}/{timestamp}", routes.RPointHandler)
	r.HandleFunc("/api/publication/{R}", routes.PublicationHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static"))))

	// CORS
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	logging.Info.Println("Listening on port 3000")

	logging.Error.Fatal(http.ListenAndServe(":3000", handlers.CORS(originsOk, headersOk, methodsOk)(logging.WebLoggingMiddleware(r))))
}
