package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/goauth2/oauth"
	"github.com/golang/oauth2"
	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
)

const (
	CODE_REDIRECT = 302
	KEY_TOKEN     = "oauth2_token"
)

type User struct {
	Fullname string
	Avatar   string
}

type OAuth struct {
	PathLogin    string // Path to handle OAuth 2.0 logins.
	PathLogout   string // Path to handle OAuth 2.0 logouts.
	PathCallback string // Path to handle callback from OAuth 2.0 backend
	PathError    string // Path to handle error cases.
	OAuthConfig  *oauth2.Config
	Config       *config
	users        map[string]*User
	store        sessions.Store
	sync.Mutex
}

func NewOAuth(config *config) *OAuth {
	authUrl := "https://github.com/login/oauth/authorize"
	tokenUrl := "https://github.com/login/oauth/access_token"

	opts := &oauth2.Options{
		ClientID:     config.HTTP.GithubID,
		ClientSecret: config.HTTP.GithubSecret,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{"read:org"},
	}

	oauthConfig, err := oauth2.NewConfig(opts, authUrl, tokenUrl)
	if err != nil {
		panic(fmt.Sprintf("oauth2: %s", err))
	}

	return &OAuth{
		PathLogin:    "/login",
		PathLogout:   "/logout",
		PathCallback: "/oauth2callback",
		PathError:    "/oauth2error",
		OAuthConfig:  oauthConfig,
		Config:       config,
		users:        make(map[string]*User, 0),
		store:        sessions.NewCookieStore([]byte("something-very-secret")),
	}
}

func (o *OAuth) Handler(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "GET" {
		switch r.URL.Path {
		case o.PathLogin:
			o.HandleLogin(w, r)
			return false
		case o.PathLogout:
			o.HandleLogout(w, r)
			return false
		case o.PathCallback:
			o.HandleCallback(w, r)
			return false
		}
	}

	token := o.getToken(r)
	failed := true
	if token != nil && !token.Expired() {
		if user, err := o.getUser(token); err == nil {
			fmt.Println("TODO", user)
			failed = false
		} else {
			w.Write([]byte(err.Error()))
			return false
		}

	}

	if failed {
		next := url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, o.PathLogin+"?next="+next, CODE_REDIRECT)
		//session.Delete(KEY_TOKEN)
		return false
	}

	return true
}

func (o *OAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")

	t, err := o.OAuthConfig.NewTransportWithCode(code)
	if err != nil {
		// Pass the error message, or allow dev to provide its own
		// error handler.
		http.Redirect(w, r, o.PathError, CODE_REDIRECT)
		return
	}

	o.setToken(w, r, t.Token())
	http.Redirect(w, r, next, CODE_REDIRECT)
}

func (o *OAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	if o.getToken(r) == nil {
		// User is not logged in.
		if next == "" {
			next = "/"
		}
		http.Redirect(w, r, o.OAuthConfig.AuthCodeURL(next, "", ""), CODE_REDIRECT)
		return
	}

	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, CODE_REDIRECT)
}

func (o *OAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	//s.Delete(KEY_TOKEN)
	http.Redirect(w, r, next, CODE_REDIRECT)
}

func (s *OAuth) getToken(r *http.Request) *oauth2.Token {
	session, _ := s.store.Get(r, KEY_TOKEN)
	if raw, ok := session.Values["token"]; !ok {
		return nil
	} else {
		var tk oauth2.Token
		json.Unmarshal(raw.([]byte), &tk)

		return &tk
	}
}

func (s *OAuth) setToken(w http.ResponseWriter, r *http.Request, t *oauth2.Token) {
	session, _ := s.store.Get(r, KEY_TOKEN)
	val, _ := json.Marshal(t)
	session.Values["token"] = val
	session.Save(r, w)
}

func (o *OAuth) getUser(token *oauth2.Token) (*User, error) {
	o.Lock()
	user, ok := o.users[token.AccessToken]
	o.Unlock()

	if ok {
		return user, nil
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: token.AccessToken},
	}

	c := github.NewClient(t.Client())
	guser, _, err := c.Users.Get("")
	if err != nil {
		return nil, err
	}

	if org := o.Config.HTTP.GithubOrganization; org != "" {
		m, _, err := c.Organizations.IsMember(org, *guser.Login)
		if err != nil {
			return nil, err
		}

		if !m {
			return nil, errors.New(fmt.Sprintf(
				"User %q should be member of %q", *guser.Login, org,
			))
		}
	}

	user = &User{Fullname: *guser.Name, Avatar: *guser.AvatarURL}

	o.Lock()
	o.users[token.AccessToken] = user
	o.Unlock()

	return user, nil
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
