package main

import (
	"log"
	"net/http"
)

func authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	log.Println("Auth Middleware")

	session, err := store.Get(r, "lizzard")
	if err != nil {
		log.Println("Error getting cookie")
		log.Println(err.Error())
	}

	name := session.Values["name"]
	email := session.Values["email"]
	token := session.Values["accessToken"]

	if token == nil {
		log.Println("No token")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	log.Printf("%s, %s, %s", name, email, token)
}
