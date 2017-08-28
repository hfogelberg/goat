package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hfogelberg/goat"
	// "github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/urfave/negroni"
)

var (
	db         *sql.DB
	HmacSecret = "secret"
	store      *sessions.CookieStore
)

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://Henrik:password@localhost/helenart?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	router := mux.NewRouter().StrictSlash(true)

	goat.New(db, sessions.NewCookieStore([]byte(HmacSecret)), "/public", "/admin", "lizzardcookie", true)

	// Routing
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/login", goat.GoogleLoginHandler)
	router.HandleFunc("/callback", goat.CallbackHandler)
	router.HandleFunc("/admin", adminHandler)
	router.HandleFunc("/public", publicHandler)

	adm := router.PathPrefix("/admin").Subrouter()
	adm.HandleFunc("/", adminHandler)

	mux := http.NewServeMux()
	mux.Handle("/", router)
	mux.Handle("/admin/", negroni.New(
		negroni.HandlerFunc(goat.AuthMiddleware),
		negroni.Wrap(router),
	))

	static := http.StripPrefix("/public/", http.FileServer(http.Dir("public")))
	router.PathPrefix("/public").Handler(static)

	n := negroni.Classic()
	n.UseHandler(mux)

	n.Run(":3000")
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

func publicHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.New("").ParseFiles("templates/public.html", "templates/layout.html")
	err = tpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Printf("Error serving Admin %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
