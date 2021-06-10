package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"strconv"
	"strings"
	_ "strings"
)

func main() {

	a := App{}
	a.Initialize(
		"postgres",
		"TestPassw0rd!",
		"db")
	esapi.DeleteRequest{
		Index: "test",
	}.Do(context.Background(), a.ESC)

	res1, err := a.ESC.Index("test", nil)
	if err != nil {
		panic(err)
	}
	log.Print(res1)

	a.createITable()
	var docs []string
	if posts, err := readcsv("/app/posts.csv"); err == nil {
		for _, post := range posts {
			p := &postdb{
				ID:          uuid.New(),
				CreatedDate: post.CreatedDate,
				Rubrics:     post.Rubrics,
				Text:        post.Text,
			}
			doc1 := ElasticDocs{}
			doc1.SomeStr = p.Text
			doc1.SomeUUID = p.ID.String()

			docStr1 := jsonStruct(doc1)
			if err != nil {
				panic(err)
			}

			if err := p.createPost(a.DB); err != nil {
				panic(err)
			}
			docs = append(docs, docStr1)
		}
		for i, bod := range docs {
			fmt.Println("\nDOC _id:", i+1)
			fmt.Println(bod)

			// Instantiate a request object
			req := esapi.IndexRequest{
				Index:      "test",
				DocumentID: strconv.Itoa(i + 1),
				Body:       strings.NewReader(bod),
				Refresh:    "true",
			}
			fmt.Println(reflect.TypeOf(req))
			res, err := req.Do(context.Background(), a.ESC)
			if err != nil {
				log.Fatalf("IndexRequest ERROR: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				restext := res.String()
				log.Printf("%s ERROR indexing document ID=%d", restext, i+1)
			} else {

				// Deserialize the response into a map.
				var resMap map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
					log.Printf("Error parsing the response body: %s", err)
				}
			}
		}
	}

	a.Run(":8010")
}
