package models

import (
	"github.com/gocarina/gocsv"
	"os"
)

type CSVReader struct {
	FilePath string
}

type CSVPost struct {
	CreatedDate string `csv:"created_date"`
	Rubrics     string `csv:"rubrics"`
	Text        string `csv:"text"`
}

func (reader *CSVReader) Initialize(filepath string) {
	reader.FilePath = filepath
}

func (reader *CSVReader) ReadContent() ([]CSVPost, error) {
	in, err := os.Open(reader.FilePath)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	var posts []CSVPost

	if err := gocsv.UnmarshalFile(in, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}
