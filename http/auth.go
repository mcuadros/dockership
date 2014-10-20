package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.google.com/p/goauth2/oauth"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/golang/oauth2"
	"github.com/google/go-github/github"
	"github.com/martini-contrib/sessions"
)

const (
	authUrl      = "https://github.com/login/oauth/authorize"
	tokenUrl     = "https://github.com/login/oauth/access_token"
	codeRedirect = 302
	keyToken     = "oauth2_token"
	keyNextPage  = "next"
)

var (
	// Path to handle OAuth 2.0 logins.
	PathLogin = "/login"
	// Path to handle OAuth 2.0 logouts.
	PathLogout = "/logout"
	// Path to handle callback from OAuth 2.0 backend
	// to exchange credentials.
	PathCallback = "/oauth2callback"
	// Path to handle error cases.
	PathError = "/oauth2error"
)

type SessionHandler func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request)

func (s *server) configureAuth() {

	s.martini.Use(sessions.Sessions("my_session", sessions.NewCookieStore([]byte("foo"))))
	s.martini.Use(s.OAuth2Provider(&oauth2.Options{
		ClientID:     s.config.HTTP.GithubID,
		ClientSecret: s.config.HTTP.GithubSecret,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{"read:org"},
	}))

	s.martini.Get("/user", func(token Token, render render.Render) {
		render.JSON(200, token.GetUser())

		fmt.Println(token.Access())
	})
}

func (s *server) OAuth2Provider(opts *oauth2.Options) martini.Handler {
	config, err := oauth2.NewConfig(opts, authUrl, tokenUrl)
	if err != nil {
		panic(fmt.Sprintf("oauth2: %s", err))
	}

	return func(session sessions.Session, ctx martini.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Println(string(session.Get(keyToken).([]byte)))

		if r.Method == "GET" {
			switch r.URL.Path {
			case PathLogin:
				login(config, session, w, r)
			case PathLogout:
				logout(session, w, r)
			case PathCallback:
				handleOAuth2Callback(config, session, w, r)
			}
		}

		fmt.Println("AUTH")
		token := unmarshallToken(session)
		failed := true
		if token != nil && !token.IsExpired() {
			t := &oauth.Transport{
				Token: &oauth.Token{AccessToken: token.Access()},
			}

			c := github.NewClient(t.Client())
			if user, _, err := c.Users.Get(""); err != nil {
				panic(err)
			} else {
				token.user = user
			}

			if org := s.config.HTTP.GithubOrganization; org != "" {
				m, _, err := c.Organizations.IsMember(org, *token.user.Login)
				if err != nil {
					panic(err)
				}

				failed = !m
			} else {
				failed = false
			}

			ctx.Map(token)
		}

		if failed {
			next := url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, PathLogin+"?next="+next, codeRedirect)
			session.Delete(keyToken)
		}
	}
}

func handleOAuth2Callback(c *oauth2.Config, s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")
	t, err := c.NewTransportWithCode(code)
	if err != nil {
		// Pass the error message, or allow dev to provide its own
		// error handler.
		http.Redirect(w, r, PathError, codeRedirect)
		return
	}
	// Store the credentials in the session.
	val, _ := json.Marshal(t.Token())
	s.Set(keyToken, val)
	http.Redirect(w, r, next, codeRedirect)
}

type Token interface {
	GetUser() *github.User
	Access() string
}

type token struct {
	user *github.User
	oauth2.Token
}

func (t *token) Access() string {
	return t.AccessToken
}

func (t *token) IsExpired() bool {
	if t == nil {
		return true
	}

	return t.Expired()
}

func (t *token) GetUser() *github.User {
	return t.user
}

// Returns the refresh token.
func (t *token) Refresh() string {
	return t.RefreshToken
}

func login(c *oauth2.Config, s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	if s.Get(keyToken) == nil {
		// User is not logged in.
		if next == "" {
			next = "/"
		}
		http.Redirect(w, r, c.AuthCodeURL(next, "", ""), codeRedirect)
		return
	}
	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, codeRedirect)
}

func logout(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	s.Delete(keyToken)
	http.Redirect(w, r, next, codeRedirect)
}

func unmarshallToken(s sessions.Session) (t *token) {
	if s.Get(keyToken) == nil {
		return
	}
	data := s.Get(keyToken).([]byte)
	var tk oauth2.Token
	json.Unmarshal(data, &tk)
	return &token{Token: tk}
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
