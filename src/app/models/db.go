package models

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
)

type Database struct {
	client *sql.DB
}

type DBPost struct {
	ID          uuid.UUID
	CreatedDate string
	Rubrics     string
	Text        string
}

func (db *Database) Initialize(host, port, user, password, dbname string) error {
	connectionString :=
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db.client, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (db *Database) down() {
	db.client.QueryRow("DROP SCHEMA public CASCADE")
}

func (db *Database) up() error {

	const tableCreationQuery = `CREATE SCHEMA public; CREATE TABLE IF NOT EXISTS posts
	(
		id uuid,
		text TEXT NOT NULL,
		created_date DATE,
		rubrics TEXT,
		CONSTRAINT posts_pkey PRIMARY KEY (id)
	)`

	if _, err := db.client.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (db *Database) Recreate() error {
	db.down()
	return db.up()
}

func (db *Database) getPostsByIds(ids []uuid.UUID) ([]DBPost, error) {

	var concatIds = ""

	for _, id := range ids {
		concatIds = concatIds + "'" + id.String() + "',"
	}
	concatIds = concatIds[:len(concatIds)-1]

	query := "SELECT id, text, created_date, rubrics FROM public.posts WHERE id IN (" + concatIds + ")"

	rows, err := db.client.Query(query)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer rows.Close()

	var posts []DBPost

	for rows.Next() {
		var p DBPost
		if err := rows.Scan(&p.ID, &p.Text, &p.CreatedDate, &p.Rubrics); err != nil {
			log.Fatal(err)
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (p *DBPost) CreatePost(db *Database) error {
	err := db.client.QueryRow(
		"INSERT INTO posts(id, created_date, rubrics,text) VALUES($1, $2, $3, $4) RETURNING id",
		p.ID, p.CreatedDate, p.Rubrics, p.Text).Scan(&p.ID)

	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemovePost(id uuid.UUID) error {
	if _, err := db.client.Exec("DELETE FROM posts WHERE id=$1", id); err != nil {
		return err
	}
	return nil
}
