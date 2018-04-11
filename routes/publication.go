package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gertjaap/dlcoracle/store"

	"github.com/gertjaap/dlcoracle/datasources"
	"github.com/gertjaap/dlcoracle/logging"

	"github.com/gorilla/mux"
)

type PublicationResponse struct {
	Value     uint64 `json:"value"`
	Signature string `json:"signature"`
}

func PublicationHandler(w http.ResponseWriter, r *http.Request) {
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

	value, signature, err := store.GetPublication(datasourceId, timestamp)
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
