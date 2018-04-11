package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gertjaap/dlcoracle/store"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/datasources"
	"github.com/gertjaap/dlcoracle/logging"

	"github.com/gorilla/mux"
)

type SubscribeResponse struct {
	Success bool   `json:"success"`
	PubKey  string `json:"pubKey"`
}

func SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	datasourceId, err := strconv.ParseUint(vars["datasource"], 10, 64)
	if err != nil {
		logging.Error.Println("SubscribeHandler - Cannot parse Datasource: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !datasources.HasDatasource(datasourceId) {
		logging.Error.Println("SubscribeHandler - Invalid Datasource: ", datasourceId)
		http.Error(w, fmt.Sprintf("Invalid datasource %d", datasourceId), http.StatusInternalServerError)
		return
	}

	timestamp, err := strconv.ParseUint(vars["timestamp"], 10, 64)
	if err != nil {
		logging.Error.Println("SubscribeHandler - Cannot parse Timestamp: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rPoint, err := crypto.DeriveR(datasourceId, timestamp)
	if err != nil {
		logging.Error.Println("SubscribeHandler - Cannot derive R point: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = store.Subscribe(datasourceId, timestamp)
	if err != nil {
		logging.Error.Println("SubscribeHandler - Could not subscribe:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SubscribeResponse{Success: true, PubKey: hex.EncodeToString(rPoint[:])}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
