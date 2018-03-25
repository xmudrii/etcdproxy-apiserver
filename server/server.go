package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/xmudrii/etcdproxy-apiserver/etcdapi"
	"github.com/xmudrii/etcdproxy-apiserver/pkg/env"
)

type key struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

var (
	APIServerPort string
)

// RunServer runs mux server for external-API.
func RunServer() {
	APIServerPort = env.GetEnvString("ETCDAPISERVER_PORT", "8000")
	APIServerPort = ":" + APIServerPort

	router := mux.NewRouter()
	router.HandleFunc("/write", WriteKey).Methods("PUT")
	log.Fatal(http.ListenAndServe(APIServerPort, router))
}

// WriteKey writes to etcd.
func WriteKey(w http.ResponseWriter, r *http.Request) {
	var k key
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&k); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := etcdapi.WriteEtcd("http://127.0.0.1:23790", k.Name, k.Value); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, k)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
