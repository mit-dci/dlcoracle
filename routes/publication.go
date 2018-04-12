package routes

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gertjaap/dlcoracle/store"

	"github.com/gertjaap/dlcoracle/logging"

	"github.com/gorilla/mux"
)

type PublicationResponse struct {
	Value     uint64 `json:"value"`
	Signature string `json:"signature"`
}

func PublicationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	passedR, err := hex.DecodeString(vars["R"])
	if err != nil {
		logging.Error.Println("SubscribeHandler - Error parsing R: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var R [33]byte
	copy(R[:], passedR[:])

	value, signature, err := store.GetPublication(R)
	if err != nil {
		logging.Error.Println("SubscribeHandler - Error getting publication: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PublicationResponse{Value: value, Signature: hex.EncodeToString(signature[:])}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
