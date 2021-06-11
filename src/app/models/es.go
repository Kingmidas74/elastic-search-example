package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	es "github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"strconv"
	"strings"
)

type ESSettings struct {
	Host string
	Port int
}

type Elastic struct {
	Settings ESSettings
	Client   *es.Client
}

type ESDocument struct {
	Content string
}

func (doc *ESDocument) ToJSON() (string, error) {

	docStruct := &ESDocument{
		Content: doc.Content,
	}

	b, err := json.Marshal(docStruct)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (elastic *Elastic) Initialize(settings ESSettings) error {
	cf := es.Config{
		Addresses: []string{
			"http://" + settings.Host + ":" + strconv.Itoa(settings.Port),
		},
	}
	e, err := es.NewClient(cf)
	if err != nil {
		return err
	}
	elastic.Client = e
	return nil
}

func (elastic *Elastic) CreateIndex(index string) error {

	var indices []string
	indices = append(indices, index)
	_, err := elastic.Client.DeleteByQuery(indices, nil)
	if err != nil {
		return err
	}

	if _, err := elastic.Client.Index(index, nil); err != nil {
		return err
	}
	return nil
}

func (elastic *Elastic) UploadDocument(index string, id string, doc ESDocument) (bool, error) {
	jsonDoc, err := doc.ToJSON()
	if err != nil {
		return false, err
	}
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       strings.NewReader(jsonDoc),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), elastic.Client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return false, errors.New(res.String())
	}

	var resMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
		return false, err
	}
	return true, nil
}

func (elastic *Elastic) DropDocument(index string, id string) error {

	response, err := elastic.Client.Delete(index, id)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

func (elastic *Elastic) SearchDocuments(index string, key string, term string) (map[string]string, error) {
	q := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				key: term},
		},
	}

	var bts bytes.Buffer
	if err := json.NewEncoder(&bts).Encode(q); err != nil {
		return nil, err
	}

	response, err := elastic.Client.Search(
		elastic.Client.Search.WithContext(context.Background()),
		elastic.Client.Search.WithIndex(index),
		elastic.Client.Search.WithBody(&bts),
		elastic.Client.Search.WithTrackTotalHits(true),
		elastic.Client.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var elasticResult map[string]interface{}
	var result = map[string]string{}
	json.NewDecoder(response.Body).Decode(&elasticResult)
	for _, hit := range elasticResult["hits"].(map[string]interface{})["hits"].([]interface{}) {
		doc := hit.(map[string]interface{})["_source"].(map[string]interface{})

		var id = hit.(map[string]interface{})["_id"].(string)
		var content = doc["Content"].(string)

		result[id] = content
	}

	return result, nil
}
