package main

import (
	"github.com/gocarina/gocsv"
	"log"
	"os"
)

func readcsv(filePath string) ([]postcsv, error) {
	in, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer in.Close()

	posts := []postcsv{}

	if err := gocsv.UnmarshalFile(in, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}
