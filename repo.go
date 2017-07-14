package main

import (
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

func createToken(userID int, token string) error {
	sql := "INSERT INTO user_tokens (user_id, token) VALUES ($1, $2)"
	_, err := db.Exec(sql, userID, token)
	if err != nil {
		return err
	}
	return nil
}
