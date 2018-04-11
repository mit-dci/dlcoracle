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
	crypto.StoreKey(key)
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
	r.HandleFunc("/api/pubkey", routes.PubKeyHandler)
	r.HandleFunc("/api/rpointpubkey/{datasource}/{timestamp}", routes.RPointPubKeyHandler)
	r.HandleFunc("/api/subscribe/{datasource}/{timestamp}", routes.SubscribeHandler)
	r.HandleFunc("/api/publication/{datasource}/{timestamp}", routes.PublicationHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static"))))

	logging.Info.Println("Listening on port 3000")

	logging.Error.Fatal(http.ListenAndServe(":3000", logging.WebLoggingMiddleware(r)))
}
