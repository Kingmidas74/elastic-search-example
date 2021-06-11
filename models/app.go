package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *Database
	ESC    *Elastic
}

const TestIndex = "test"

func (a *App) Initialize(db *Database, elastic *Elastic) {
	a.DB = db
	a.ESC = elastic
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(port int) {
	err := http.ListenAndServe(":"+strconv.Itoa(port), a.Router)
	if err != nil {
		log.Fatal(err)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/posts/{token:.+}", a.getPostsByToken).Methods("GET")
}

func (a *App) getPostsByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	var uuids []uuid.UUID

	if docs, err := a.ESC.SearchDocuments(TestIndex, "Content", token); err == nil {
		for id, _ := range docs {
			uuids = append(uuids, uuid.MustParse(id))
		}
	} else {
		respondWithError(w, 400, err.Error())
	}

	if posts, err := a.DB.getPostsByIds(uuids); err == nil {
		respondWithJSON(w, http.StatusOK, posts)
	} else {
		respondWithError(w, 400, err.Error())
	}
}
