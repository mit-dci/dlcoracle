package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/datasources"
	"github.com/gertjaap/dlcoracle/logging"

	"github.com/gorilla/mux"
)

type PubKeyResponse struct {
	PubKey string `json:"pubKey"`
}

func PubKeyHandler(w http.ResponseWriter, r *http.Request) {
	pubKey, err := crypto.GetPubKey()
	if err != nil {
		logging.Error.Println("PubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	returnPubKey(w, pubKey[:])

}

// TODO: Should only return if a subscription exists for people re-requesting the R point. That should prevent people using an unregistered R point (which will never be published)
func RPointPubKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	datasourceId, err := strconv.ParseUint(vars["datasource"], 10, 64)
	if err != nil {
		logging.Error.Println("RPointPubKeyHandler - Invalid Datasource: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !datasources.HasDatasource(datasourceId) {
		logging.Error.Println("RPointPubKeyHandler - Invalid Datasource: ", datasourceId)
		http.Error(w, fmt.Sprintf("Invalid datasource %d", datasourceId), http.StatusInternalServerError)
		return
	}

	timestamp, err := strconv.ParseUint(vars["timestamp"], 10, 64)
	if err != nil {
		logging.Error.Println("RPointPubKeyHandler - Invalid Timestamp: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rPoint, err := crypto.DeriveR(datasourceId, timestamp)
	if err != nil {
		logging.Error.Println("RPointPubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	returnPubKey(w, rPoint[:])
}

func returnPubKey(w http.ResponseWriter, pubKey []byte) {
	response := PubKeyResponse{PubKey: hex.EncodeToString(pubKey)}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
