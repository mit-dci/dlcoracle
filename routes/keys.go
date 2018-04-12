package routes

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/logging"
)

type PubKeyResponse struct {
	A string
	B string
	Q string
}

func PubKeyHandler(w http.ResponseWriter, r *http.Request) {
	A, err := crypto.GetPubKey(crypto.KeyTypeA)
	if err != nil {
		logging.Error.Println("PubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	B, err := crypto.GetPubKey(crypto.KeyTypeB)
	if err != nil {
		logging.Error.Println("PubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Q, err := crypto.GetPubKey(crypto.KeyTypeQ)
	if err != nil {
		logging.Error.Println("PubKeyHandler", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PubKeyResponse{
		A: hex.EncodeToString(A[:]),
		B: hex.EncodeToString(B[:]),
		Q: hex.EncodeToString(Q[:])}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
