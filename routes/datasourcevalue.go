package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/gertjaap/dlcoracle/datasources"
)

type DataSourceValueResponse struct {
	CurrentValue uint64 `json:"currentValue"`
	ValueError   string `json:"valueError,omitempty"`
}

func DataSourceValueHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	response := DataSourceValueResponse{}
	datasourceId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		response.ValueError = err.Error()
	}
	ds, err := datasources.GetDatasource(datasourceId)

	response.CurrentValue, err = ds.Value()
	if err != nil {
		response.ValueError = err.Error()
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
