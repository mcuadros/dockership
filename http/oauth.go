package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/mcuadros/dockership/config"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
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
	Config       *config.Config
	users        map[string]*User
	store        sessions.Store
	sync.Mutex
}

func NewOAuth(config *config.Config) *OAuth {
	authUrl := "https://github.com/login/oauth/authorize"
	tokenUrl := "https://github.com/login/oauth/access_token"

	return &OAuth{
		PathLogin:    "/login",
		PathLogout:   "/logout",
		PathCallback: "/oauth2callback",
		PathError:    "/oauth2error",
		OAuthConfig: &oauth2.Config{
			ClientID:     config.HTTP.GithubID,
			ClientSecret: config.HTTP.GithubSecret,
			Scopes:       []string{"read:org"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  authUrl,
				TokenURL: tokenUrl,
			},
		},
		Config: config,
		users:  make(map[string]*User, 0),
		store:  sessions.NewCookieStore([]byte("cookie-key")),
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
	if token != nil && token.Valid() {
		if _, err := o.getValidUser(token); err == nil {
			failed = false
		} else {
			w.Write([]byte(err.Error()))
			return false
		}

	}

	if failed {
		next := url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, o.PathLogin+"?next="+next, CODE_REDIRECT)
		return false
	}

	return true
}

func (o *OAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")

	tok, err := o.OAuthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		// Pass the error message, or allow dev to provide its own
		// error handler.
		http.Redirect(w, r, o.PathError, CODE_REDIRECT)
		return
	}

	o.setToken(w, r, tok)
	http.Redirect(w, r, next, CODE_REDIRECT)
}

func (s *OAuth) setToken(w http.ResponseWriter, r *http.Request, t *oauth2.Token) {
	session, _ := s.store.Get(r, KEY_TOKEN)
	val, _ := json.Marshal(t)
	session.Values["token"] = val
	session.Save(r, w)
}

func (o *OAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	if o.getToken(r) == nil {
		// User is not logged in.
		if next == "" {
			next = "/"
		}
		http.Redirect(w, r, o.OAuthConfig.AuthCodeURL(next), CODE_REDIRECT)
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
		var tok oauth2.Token
		json.Unmarshal(raw.([]byte), &tok)

		return &tok
	}
}

func (o *OAuth) getValidUser(token *oauth2.Token) (*User, error) {
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

	if err := o.isValidUser(c, guser); err != nil {
		return nil, err
	}

	user = &User{}
	if guser != nil && guser.Name != nil {
		user.Fullname = *guser.Name
	} else if guser.Login != nil {
		user.Fullname = *guser.Login
	}

	if guser != nil && guser.AvatarURL != nil {
		user.Avatar = *guser.AvatarURL
	}

	o.Lock()
	o.users[token.AccessToken] = user
	o.Unlock()

	return user, nil
}

func (o *OAuth) isValidUser(c *github.Client, u *github.User) error {
	if err := o.validateGithubOrganization(c, u); err != nil {
		return err
	}

	if err := o.validateGithubUser(c, u); err != nil {
		return err
	}

	return nil
}

func (o *OAuth) validateGithubOrganization(c *github.Client, u *github.User) error {
	org := o.Config.HTTP.GithubOrganization
	if org == "" {
		return nil
	}

	m, _, err := c.Organizations.IsMember(org, *u.Login)
	if err != nil {
		return err
	}

	if !m {
		return errors.New(fmt.Sprintf(
			"User %q should be member of %q", *u.Login, org,
		))
	}

	return nil
}

func (o *OAuth) validateGithubUser(c *github.Client, u *github.User) error {
	if len(o.Config.HTTP.GithubUsers) == 0 {
		return nil
	}

	for _, user := range o.Config.HTTP.GithubUsers {
		if user == *u.Login {
			return nil
		}
	}

	return errors.New(fmt.Sprintf(
		"User %q not allowed, not in the access list.", *u.Login,
	))
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
