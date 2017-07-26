package gobi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

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

	// err = saveUserToDb(user.Email, token.AccessToken)
	userID, err := createUser(user.Name, user.Email)
	if err != nil {
		fmt.Println("Erro saving user to Db")
		fmt.Println(err.Error())
		return
	}

	err = createToken(userID, token.AccessToken)
	if err != nil {
		fmt.Println("Error creating token")
		fmt.Println(err.Error())
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
