package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		Domain:   "",       // Пустой домен позволяет работать с IP-адресами
		MaxAge:   3600 * 8, // Время жизни сессии (например, 8 часов)
		HttpOnly: true,
	}
}

func GetSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return session
}

func IsAuthenticated(r *http.Request) bool {
	session, _ := store.Get(r, "user-session")
	_, ok := session.Values["username"]
	return ok
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	session.Options.MaxAge = -1
	session.Save(r, w) // Убедитесь, что сохраняете изменения
}
