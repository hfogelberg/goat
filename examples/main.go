package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hfogelberg/goat"
	"github.com/urfave/negroni"
)

const (
	HmacSecret = "secret"
	CookieName = "lizzard"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)

	goat.New(sessions.NewCookieStore([]byte(HmacSecret)), "/user", CookieName)

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/googlelogin", goat.GoogleLoginHandler)
	router.HandleFunc("/gcallback", goat.GoogleCallbackHandler)
	router.HandleFunc("/user", userHandler)

	// Serve assets
	static := http.StripPrefix("/public/", http.FileServer(http.Dir("public")))
	router.PathPrefix("/public/").Handler(static)

	n := negroni.Classic()
	n.UseHandler(router)
	port := getEnv("PORT", ":3000")
	n.Run(port)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.New("").ParseFiles("templates/index.html", "templates/layout.html")
	err = tpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Printf("Error serving Index %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	user := goat.GetGoogleUserInfo(w, r)
	fmt.Println(user.GivenName)
	fmt.Println(user.FamilyName)
	fmt.Println(user.Email)
	fmt.Println(user.AccessToken)
	fmt.Println(user.Picture)

	tpl, err := template.New("").ParseFiles("templates/user.html", "templates/layout.html")
	err = tpl.ExecuteTemplate(w, "layout", user)
	if err != nil {
		log.Printf("Error serving Admin %s\n", err.Error())
		return
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
