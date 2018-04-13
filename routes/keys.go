package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gertjaap/dlcoracle/store"
	"github.com/gorilla/mux"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/datasources"
	"github.com/gertjaap/dlcoracle/logging"
)

type RPointResponse struct {
	R string
}

func RPointHandler(w http.ResponseWriter, r *http.Request) {
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

	rPoint, err := store.GetRPoint(datasourceId, timestamp)
	if err != nil {
		logging.Error.Println("RPointPubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := RPointResponse{
		R: hex.EncodeToString(rPoint[:]),
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type PubKeyResponse struct {
	A string
}

func PubKeyHandler(w http.ResponseWriter, r *http.Request) {
	A, err := crypto.GetPubKey(crypto.KeyTypeA)
	if err != nil {
		logging.Error.Println("PubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PubKeyResponse{
		A: hex.EncodeToString(A[:]),
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
