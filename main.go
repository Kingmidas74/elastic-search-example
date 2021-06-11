package main

import (
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"main/models"
	"os"
	"strconv"
	_ "strings"
)

func main() {

	db := models.Database{}
	if err := db.Initialize(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")); err != nil {
		log.Fatal(err)
	} else {
		err := db.Recreate()
		if err != nil {
			log.Fatal(err)
		}
	}

	elastic := models.Elastic{}
	if esPort, err := strconv.Atoi(os.Getenv("ES_PORT")); err == nil {
		err := elastic.Initialize(models.ESSettings{
			Host: os.Getenv("ES_HOST"),
			Port: esPort,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	a := models.App{}
	a.Initialize(&db, &elastic)

	Initialize(&a)

	a.Run(8010)
}

func Initialize(app *models.App) {
	csvReader := models.CSVReader{}
	csvReader.Initialize(os.Getenv("SEED_FILE_PATH"))
	if csvPosts, err := csvReader.ReadContent(); err == nil {
		if err := app.ESC.CreateIndex(models.TestIndex); err != nil {
			log.Fatal(err)
		}

		for _, csvPost := range csvPosts {
			dbPost := models.DBPost{}
			dbPost.ID = uuid.New()
			dbPost.CreatedDate = csvPost.CreatedDate
			dbPost.Rubrics = csvPost.Rubrics
			dbPost.Text = csvPost.Text

			if err := dbPost.CreatePost(app.DB); err != nil {
				log.Fatal(err)
			}
			esDocument := models.ESDocument{}
			esDocument.Content = csvPost.Text

			uploadResult, err := app.ESC.UploadDocument(models.TestIndex, dbPost.ID.String(), esDocument)
			if err != nil {
				log.Fatal(err)
			}
			if uploadResult == false {
				log.Fatalf("Couldn't upload document %s", dbPost.ID.String())
			}
		}
	}
}
