package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func createUser(name string, email string) (int, error) {
	sql := "INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id"
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err.Error())
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(email, name).Scan(&id)
	if err != nil {
		log.Fatal(err.Error())
		return 0, err
	}

	log.Println("Last insert id")
	log.Println(id)

	return id, nil
}

func getUserIdByEmail(email string) (int, error) {
	id := 0
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		log.Println("No user with that email")
		return 0, nil
	case err != nil:
		log.Println(err.Error())
		return 0, err
	default:
		log.Println("Email is in Db")
		return id, nil
	}
}

func createToken(userID int, token string) error {
	sql := "INSERT INTO user_tokens (user_id, token) VALUES ($1, $2)"
	_, err := db.Exec(sql, userID, token)
	if err != nil {
		return err
	}
	return nil
}

func tokenIsValid(token string) (bool, error) {
	log.Println("Checking if token is Valid")
	id := 0
	err := db.QueryRow("SELECT user_id from user_tokens WHERE token = $1", token).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		log.Println("No user with that token.")
		return false, nil
	case err != nil:
		log.Println(err.Error())
		return false, err
	default:
		log.Println("Token is Ok!")
		return true, nil
	}
}
