package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PostGreClient struct {
	db *sql.DB
}

func NewPostGreClient(db *sql.DB) *PostGreClient {
	return &PostGreClient{db: db}
}

func (n *PostGreClient) Connect() error {
	pbDB, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=your_password dbname=your_database sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pbDB.Close()
	err = pbDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ket noi thanh cong")
	return nil
}
