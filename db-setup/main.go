package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host = "postgre-postgresql.postgre.svc.cluster.local"
	// host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "rootpassword"
	dbname   = "postgres"
)

const (
	createRoleStmt = `
DROP ROLE IF EXISTS	"ro";
CREATE ROLE "ro" NOINHERIT;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO "ro";`
	createTableStmt = `
DROP TABLE IF EXISTS public.users;
CREATE TABLE IF NOT EXISTS public.users (
	id SERIAL PRIMARY KEY,
	age INT,
	first_name TEXT,
	last_name TEXT,
	email TEXT UNIQUE NOT NULL
);`
	createUserStmt = `
INSERT INTO public.users (age, email, first_name, last_name)
VALUES ($1, $2, $3, $4);`
)

var users = []struct {
	age                          int
	first_name, last_name, email string
}{
	{
		age:        30,
		first_name: "Jonathan",
		last_name:  "Christopher",
		email:      "example@hotmail.com",
	},
	{
		age:        42,
		first_name: "Alan",
		last_name:  "Turing",
		email:      "crack@enigma.co.uk",
	},
	{
		age:        13,
		first_name: "Mister",
		last_name:  "Droelf",
		email:      "drei@zwoelf.de",
	},
	{
		age:        13,
		first_name: "Lisa",
		last_name:  "Haus",
		email:      "lisa@haus.de",
	},
	{
		age:        33,
		first_name: "New",
		last_name:  "from K8s",
		email:      "new@k8s.de",
	},
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	log.Print("opening db connection")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}
	defer db.Close()

	log.Print("pinging db")
	err = db.Ping()
	if err != nil {
		log.Fatal("error pinging database: ", err)
	}
	log.Print("successfully connected!")

	log.Print("creating roles")
	_, err = db.Exec(createRoleStmt)
	if err != nil {
		log.Fatal("error when creating role: ", err)
	}

	log.Println("creating sample table")
	_, err = db.Exec(createTableStmt)
	if err != nil {
		log.Fatal("error when creating table: ", err)
	}

	log.Println("creating sample table entries")
	for _, u := range users {
		_, err = db.Exec(
			createUserStmt,
			u.age,
			u.email,
			u.first_name,
			u.last_name,
		)
		if err != nil {
			log.Fatal("error when creating user: ", err)
		}
	}

	log.Print("creation successful")
}
