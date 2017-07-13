package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:3000/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	// Some random string, random for each request
	oauthStateString = "random"
	hmacSampleSecret = []byte(HmacSecret)
	store            *sessions.CookieStore
	db               *sql.DB
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

const (
	HmacSecret = "secret"
	Port       = ":3000"
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
	store = sessions.NewCookieStore([]byte(HmacSecret))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/googlelogin", authHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/admin", adminHandler)

	http.ListenAndServe(Port, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	s := `<html><body><a href="/googlelogin">Log in with Google</a></body></html>`
	fmt.Fprintf(w, s)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("Code exchange failed with error ", err.Error())
		return
	}

	if !token.Valid() {
		fmt.Println("Retreived invalid token")
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("Error getting user from token ", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	var user *GoogleUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		fmt.Println("Error unmarshaling Google user", err.Error())
		return
	}

	sql := "INSERT INTO users(email, token) VALUES ($1, $2)"
	_, err = db.Exec(sql, user.Email, token.AccessToken)
	if err != nil {
		fmt.Println("Error inserting user", err.Error())
		return
	}

	session, err := store.Get(r, "lizzard")
	if err != nil {
		fmt.Println("Error getting session", err.Error())
		return
	}

	session.Values["name"] = user.Name
	session.Values["email"] = user.Email
	session.Values["picture"] = user.Picture
	session.Values["accessToken"] = token.AccessToken
	session.Save(r, w)

	http.Redirect(w, r, "/admin", 302)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Admin")

	session, _ := store.Get(r, "lizzard")

	fmt.Fprintln(w, "Name: ", session.Values["name"])
	fmt.Fprintln(w, "Email: ", session.Values["email"])
	fmt.Fprintln(w, "Token: ", session.Values["accessToken"])
}
