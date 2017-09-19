package main

import (
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

	goat.New(sessions.NewCookieStore([]byte(HmacSecret)), "/admin", CookieName)

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/googlelogin", goat.GoogleLoginHandler)
	router.HandleFunc("/callback", goat.GoogleCallbackHandler)
	router.HandleFunc("/admin", adminHandler)

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

func adminHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.New("").ParseFiles("templates/admin.html", "templates/layout.html")
	err = tpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Printf("Error serving Admin %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
