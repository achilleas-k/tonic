package tonic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gogs/go-gogs-client"
	"github.com/gorilla/mux"
)

var tokens map[string][]*gogs.AccessToken

func mockGIN() *http.Server {
	router := new(mux.Router)

	router.HandleFunc("/users/{username}/tokens", getTokens).Methods("GET")
	router.HandleFunc("/users/{username}/tokens", addToken).Methods("POST")

	tokens = make(map[string][]*gogs.AccessToken)

	srv := new(http.Server)
	srv.Handler = router
	return srv
}

func getTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["username"]
	userTokens := tokens[user]
	writeResponse(w, http.StatusOK, userTokens)
}

func addToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["username"]
	dataRaw, _ := ioutil.ReadAll(r.Body)
	data := new(gogs.CreateAccessTokenOption)
	json.Unmarshal(dataRaw, data)
	newToken := &gogs.AccessToken{Name: data.Name, Sha1: fmt.Sprintf("%s-token", data.Name)}
	if tokens[user] == nil {
		tokens[user] = make([]*gogs.AccessToken, 0, 1)
	}
	tokens[user] = append(tokens[user], newToken)
	log.Printf("Token name %q", data.Name)

	writeResponse(w, http.StatusCreated, newToken)
}

func readData(r *http.Request, obj interface{}) {
	dataRaw, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(dataRaw, obj)
}

func writeResponse(w http.ResponseWriter, status int, data interface{}) {
	dataRaw, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(dataRaw)
}
