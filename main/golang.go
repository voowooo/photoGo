package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type Profile struct {
	Id                                      uint16
	Username, Password, Photos, Description string
}

type PageData struct {
	Users        []Profile
	LoggedUserId int
}

var users = []Profile{}

var showUserPage = Profile{}

// var showUser = Profile{}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	session := GetSession(w, r)

	loggedUserID := 0
	if values, ok := session.Values["user_id"]; ok {
		if loggedUserId, ok := values.(uint16); ok {
			loggedUserID = int(loggedUserId)
		}
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query("SELECT * FROM `all_users`")
	if err != nil {
		panic(err)
	}

	users := []Profile{}
	for res.Next() {
		var user Profile
		err = res.Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	data := PageData{
		Users:        users,
		LoggedUserId: loggedUserID,
	}

	t.ExecuteTemplate(w, "index", data)
}

func register(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "register", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		var dbUser Profile
		err = db.QueryRow("SELECT * FROM `all_users` WHERE `name` = ? AND `password` = ?", username, password).Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.Photos)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session := GetSession(w, r)
		session.Values["user_id"] = dbUser.Id
		session.Values["username"] = dbUser.Username
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	t.ExecuteTemplate(w, "login", nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	ClearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func reg_user(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	password := r.FormValue("password")
	description := r.FormValue("description")

	// username = "test"

	fmt.Println(username + password)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `all_users` (`name`, `password`, `photos`, `description`) VALUES ('%s', '%s', '0', '%s')", username, password, description))
	if err != nil {
		panic(err)
	}
	defer insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func user_profile(w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)
	currentUserID, loggedIn := session.Values["user_id"].(uint16)

	loggedUserID := 0
	if values, ok := session.Values["user_id"]; ok {
		if loggedUserId, ok := values.(uint16); ok {
			loggedUserID = int(loggedUserId)
		}
	}

	vars := mux.Vars(r)
	requestedUserID := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isOwner := loggedIn && currentUserID == user.Id

	data := struct {
		Profile
		IsOwner      bool
		LoggedUserId int
	}{
		Profile:      user,
		IsOwner:      isOwner,
		LoggedUserId: loggedUserID,
	}

	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "profile", data)
}

func logUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var dbUser Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `name` = ? AND `password` = ?", username, password).Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.Photos, &dbUser.Description)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session := GetSession(w, r)
	session.Values["user_id"] = dbUser.Id
	session.Values["username"] = dbUser.Username
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func user_settings(w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)
	currentUserID, loggedIn := session.Values["user_id"].(uint16)

	loggedUserID := 0
	if values, ok := session.Values["user_id"]; ok {
		if loggedUserId, ok := values.(uint16); ok {
			loggedUserID = int(loggedUserId)
		}
	}

	vars := mux.Vars(r)
	requestedUserID := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isOwner := loggedIn && currentUserID == user.Id

	data := struct {
		Profile
		IsOwner      bool
		LoggedUserId int
	}{
		Profile:      user,
		IsOwner:      isOwner,
		LoggedUserId: loggedUserID,
	}

	t, err := template.ParseFiles("templates/settings.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "settings", data)

}

func change_desc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	description := r.FormValue("description")
	Id := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Use a prepared statement to prevent SQL injection
	stmt, err := db.Prepare("UPDATE `all_users` SET `description` = ? WHERE `id` = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Execute the query with user-provided values
	_, err = stmt.Exec(description, Id)
	if err != nil {
		panic(err)
	}

	// Redirect to the user's profile or wherever appropriate
	http.Redirect(w, r, fmt.Sprintf("/profile/%s", Id), http.StatusSeeOther)
}

func handleFunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/register", register).Methods("GET")
	rtr.HandleFunc("/login", login).Methods("GET", "POST")
	rtr.HandleFunc("/logout", logout).Methods("GET")
	rtr.HandleFunc("/reg_user", reg_user).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/change_desc", change_desc).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}", user_profile).Methods("GET")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/settings", user_settings).Methods("GET")
	rtr.HandleFunc("/log_user", logUser).Methods("GET") // Новый маршрут

	http.Handle("/", rtr)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.ListenAndServe(":8080", nil)
}

func main() {
	handleFunc()
}
