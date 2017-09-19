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

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	store       *sessions.CookieStore
	urlRedirect string
	hmacSecret  string
	cookieName  string
)

func New(s *sessions.CookieStore, url string, cName string) {
	urlRedirect = url
	store = s
	cookieName = cName
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Google login  handler")
	oauthStateString := uniuri.New()
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Google Callback Handler")

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

	session, err := store.Get(r, cookieName)
	if err != nil {
		fmt.Println("Error getting session", err.Error())
		return
	}

	session.Values["email"] = user.Email
	session.Values["givenName"] = user.GivenName
	session.Values["familyName"] = user.FamilyName
	session.Values["picture"] = user.Picture
	session.Values["accessToken"] = token.AccessToken
	session.Save(r, w)

	http.Redirect(w, r, urlRedirect, 302)
}

func GetGoogleUserInfo(w http.ResponseWriter, r *http.Request) GoogleUser {
	session, err := store.Get(r, cookieName)
	if err != nil {
		log.Printf("Error getting session cookie %s\n", err.Error())
		http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
	}

	var user GoogleUser
	user.Email = session.Values["email"].(string)
	user.GivenName = session.Values["givenName"].(string)
	user.FamilyName = session.Values["familyName"].(string)
	user.Picture = session.Values["picture"].(string)
	user.AccessToken = session.Values["accessToken"].(string)

	return user
}
