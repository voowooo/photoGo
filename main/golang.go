package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "index", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "register", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "login", nil)
}

func reg_user(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	password := r.FormValue("password")

	// username = "test"

	fmt.Println(username + password)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `all_users` (`name`, `password`) VALUES ('%s', '%s')", username, password))
	if err != nil {
		panic(err)
	}
	defer insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func log_user(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	password := r.FormValue("password")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var dbPassword string

	err = db.QueryRow("SELECT `password` FROM `all_users` WHERE `name` = ?", username).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// Пользователь не найден
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return
		} else {
			panic(err)
		}
	}

	if password == dbPassword {
		// Пароль верен, перенаправляем на страницу профиля
		http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
	} else {
		// Пароль неверен
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
	}
}

func profile(w http.ResponseWriter, r *http.Request) {
	// Получаем имя пользователя из URL
	requestedUsername := r.URL.Path[len("/profile/"):]

	// Здесь можно добавить проверку сессии, чтобы получить текущего пользователя
	// Для простоты предположим, что у нас есть переменная currentUser,
	// которая содержит имя текущего аутентифицированного пользователя.
	currentUser := "текущий_пользователь" // Это нужно заменить на реальную проверку сессии

	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	// Проверяем, имеет ли текущий пользователь права на изменение данных
	hasEditRights := currentUser == requestedUsername

	data := struct {
		Username      string
		HasEditRights bool
	}{
		Username:      requestedUsername,
		HasEditRights: hasEditRights,
	}

	t.ExecuteTemplate(w, "profile", data)
}

func handleFunc() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/register/", register)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/reg_user/", reg_user)
	http.HandleFunc("/log_user/", log_user)
	http.HandleFunc("/profile/", profile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleFunc()
}
