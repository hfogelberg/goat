package goat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const HmacSecret = "secret"

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:3000/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}

	store       *sessions.CookieStore
	urlRedirect string
	cookieName  string
)

func New(s *sessions.CookieStore, url string, cookie string) {
	urlRedirect = url
	store = s
	cookieName = cookie
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	oauthStateString := uniuri.New()
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Code exchange failed with error %s\n", err.Error())
		return
	}

	if !token.Valid() {
		log.Println("Retreived invalid token")
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error getting user from token %s\n", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	var user *GoogleUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		log.Printf("Error unmarshaling Google user %s\n", err.Error())
		return
	}

	log.Println(user.Email)
	// err = saveUserToDb(user.Email, token.AccessToken)
	// userID, err := createUser(user.Name, user.Email)
	// if err != nil {
	// 	fmt.Println("Erro saving user to Db")
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// err = createToken(userID, token.AccessToken)
	// if err != nil {
	// 	fmt.Println("Error creating token")
	// 	fmt.Println(err.Error())
	// }

	session, err := store.Get(r, cookieName)
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
