package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	es "github.com/elastic/go-elasticsearch"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	ESC    *es.Client
}

type ElasticDocs struct {
	SomeStr  string
	SomeUUID string
}

func jsonStruct(doc ElasticDocs) string {

	// Create struct instance of the Elasticsearch fields struct object
	docStruct := &ElasticDocs{
		SomeStr:  doc.SomeStr,
		SomeUUID: doc.SomeUUID,
	}

	fmt.Println("\ndocStruct:", docStruct)
	fmt.Println("docStruct TYPE:", reflect.TypeOf(docStruct))

	// Marshal the struct to JSON and check for errors
	b, err := json.Marshal(docStruct)
	if err != nil {
		fmt.Println("json.Marshal ERROR:", err)
		return string(err.Error())
	}
	return string(b)
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s host=db port=5432 dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.InitGo()

	a.Router = mux.NewRouter()

	// tom: this line is added after initializeRoutes is created later on
	a.initializeRoutes()
}

func (a *App) InitGo() {
	cf := es.Config{
		Addresses: []string{
			"http://es:9200",
		},
	}
	e, err := es.NewClient(cf)
	if err != nil {
		log.Println(err)
	}
	a.ESC = e
	res, err := a.ESC.Info()

	// Deserialize the response into a map.
	if err != nil {
		log.Fatalf("client.Info() ERROR:", err)
	} else {
		log.Printf("client response:", res)
	}
}

func (a *App) elaseticsearch(term string, key string) (by []uuid.UUID, err error) {
	q := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				key: term},
		},
	}
	var bts bytes.Buffer
	if err = json.NewEncoder(&bts).Encode(q); err != nil {
		return
	}
	r, err := a.ESC.Search(
		a.ESC.Search.WithContext(context.Background()),
		a.ESC.Search.WithIndex("test"),
		a.ESC.Search.WithBody(&bts),
		a.ESC.Search.WithTrackTotalHits(true),
		a.ESC.Search.WithPretty(),
		//a.ESC.Search.WithDocvalueFields("SomeUUID", "SomeStr"),
	)
	if err != nil {
		panic(err)
		return
	} else {
		log.Print("TYPE OF")
		log.Print(reflect.TypeOf(r))
	}
	defer r.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(r.Body).Decode(&result)
	var uuids []uuid.UUID
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		craft := hit.(map[string]interface{})["_source"].(map[string]interface{})

		var struuid = craft["SomeUUID"].(string)
		var parseduuid = uuid.MustParse(struuid)
		uuids = append(uuids, parseduuid)

		//dto := ElasticDocs{}

		//dto.SomeStr = craft t[0]["SomeStr"]
	}

	_, err = ioutil.ReadAll(r.Body)
	if err != nil || r.StatusCode != http.StatusOK {
		return
	}
	by = uuids
	return
}

// tom: initial version
// func (a *App) Run(addr string) { }
// improved version
func (a *App) Run(addr string) {
	http.ListenAndServe(":8010", a.Router)
}

// tom: these are added later
func (a *App) getPostsByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	uuids, err := a.elaseticsearch(token, "SomeStr")
	if err != nil {
		panic(err)
	}

	var uuidsstr = ""

	for _, us := range uuids {
		uuidsstr = uuidsstr + "'" + us.String() + "',"
	}
	uuidsstr = uuidsstr[:len(uuidsstr)-1]
	println(uuidsstr)

	query := "SELECT id, text, created_date, rubrics FROM public.posts WHERE id IN (" + uuidsstr + ")"

	rows, err := a.DB.Query(query)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	posts := []postdb{}

	for rows.Next() {
		var p postdb
		if err := rows.Scan(&p.ID, &p.Text, &p.CreatedDate, &p.Rubrics); err != nil {
			panic(err)
		}
		posts = append(posts, p)
	}

	respondWithJSON(w, http.StatusOK, posts)
}

func (a *App) createTable(w http.ResponseWriter, r *http.Request) {
	a.createITable()
}

func (a *App) createITable() {
	var err error
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s host=db port=5432 sslmode=disable", "postgres", "TestPassw0rd!", "db")
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	const tableCreationQuery = `DROP SCHEMA public CASCADE; 
CREATE SCHEMA public; CREATE TABLE IF NOT EXISTS posts
	(
		id uuid,
		text TEXT NOT NULL,
		created_date DATE,
		rubrics TEXT,
		CONSTRAINT posts_pkey PRIMARY KEY (id)
	)`

	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
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

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/posts/{token:.+}", a.getPostsByToken).Methods("GET")
	a.Router.HandleFunc("/table/delete", a.dropTable).Methods("GET")
	a.Router.HandleFunc("/table/create", a.createTable).Methods("GET")
}

type postcsv struct {
	CreatedDate string `csv:"created_date"`
	Rubrics     string `csv:"rubrics"`
	Text        string `csv:"text"`
}

type postdb struct {
	ID          uuid.UUID
	CreatedDate string
	Rubrics     string
	Text        string
}

func (p *postdb) createPost(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO posts(id, created_date, rubrics,text) VALUES($1, $2, $3, $4) RETURNING id",
		p.ID, p.CreatedDate, p.Rubrics, p.Text).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func (a *App) dropTable(w http.ResponseWriter, r *http.Request) {
	a.DB.QueryRow("DROP TABLE posts")
}
