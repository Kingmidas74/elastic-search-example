package main

import (
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	_ "main/docs"
	"main/models"
	"os"
	"strconv"
	_ "strings"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
func main() {

	currentEnvironment := os.Getenv("ENVIRONMENT")
	if len(currentEnvironment) == 0 || currentEnvironment == "Development" {
		godotenv.Load()
	}

	db := models.Database{}
	if err := db.Initialize(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")); err != nil {
		log.Fatal(err)
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
	if appPort, err := strconv.Atoi(os.Getenv("APP_PORT")); err == nil {
		log.Printf("Starting application on port %d", appPort)
		a.Run(appPort)
	} else {
		log.Fatal(err)
	}
}

func Initialize(app *models.App) {
	csvReader := models.CSVReader{}
	seedPath := os.Getenv("SEED_FILE_PATH")
	if len(seedPath) == 0 {
		return
	}
	err := app.DB.Recreate()
	if err != nil {
		log.Fatal(err)
		return
	}
	csvReader.Initialize(seedPath)
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
