package gobi

import (
	"log"
	"net/http"
)

func authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// 1. Get cookie. If there is no cookie, redirect to login
	session, err := store.Get(r, "lizzard")
	if err != nil {
		log.Println("Error getting cookie")
		log.Println(err.Error())
	}

	if session.Values["accessToken"] == nil {
		log.Println("No cookie. Redirect to login")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		next(w, r)
	}

	email := session.Values["email"].(string)
	token := session.Values["accessToken"].(string)

	// 2.  Cookie is Ok. Check if user is in Db
	userID, err := getUserIdByEmail(email)
	if err != nil {
		log.Println("Error validating email in Db")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	if userID == 0 {
		log.Println("Unknown email. Redirect to login")
		// delete cookie
		session.Options.MaxAge = -1
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		next(w, r)
	}

	// Email is Ok, check if token is in Db.
	tokenIsValid, err := tokenIsValid(token)
	if err != nil {
		log.Println("Error validaing token")
		// delete cookie
		session.Options.MaxAge = -1
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		next(w, r)
	}

	// Token is new, save in Db
	if tokenIsValid == false {
		err := createToken(userID, token)
		if err != nil {
			log.Println("Error creating token")
			// delete cookie
			session.Options.MaxAge = -1
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			next(w, r)
		}
	}

	next(w, r)
}
