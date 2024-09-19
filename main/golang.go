package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type Profile struct {
	Id, Userphoto                           uint16
	Username, Password, Photos, Description string
	UserphotoURL                            string
}

type LoggedUserStruct struct {
	Id, Userphoto                           uint16
	Username, Password, Photos, Description string
	UserphotoURL                            string
}

type PageData struct {
	Users              []Profile
	LoggedUserId       int
	UserphotoURL       string
	LoggedUserphotoURL string
}

var users = []Profile{}

var showUserPage = Profile{}

// var showUser = Profile{}

func index(w http.ResponseWriter, r *http.Request) {
	// Парсинг HTML-шаблонов
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/secondHeader.html")
	// t, err := template.ParseFiles("templatesNEW/index.html", "templatesNEW/header.html", "templatesNEW/footer.html")
	if err != nil {
		http.Error(w, "Ошибка при загрузке шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение сессии
	session := GetSession(w, r)
	currentUserID, ok := session.Values["user_id"].(uint16)

	// Подключение к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Запрос на получение всех пользователей из таблицы `all_users`
	rows, err := db.Query("SELECT id, name, password, photos, description, userphoto FROM `all_users`")
	if err != nil {
		http.Error(w, "Ошибка выполнения запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание пустого массива пользователей
	var users []Profile
	for rows.Next() {
		var user Profile
		err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto)
		if err != nil {
			http.Error(w, "Ошибка чтения данных пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Преобразование ID фотографии в URL для отображения
		if user.Userphoto != 0 {
			user.UserphotoURL = "/userphoto/" + strconv.Itoa(int(user.Userphoto))
		} else {
			user.UserphotoURL = "/static/icons/profile_page-icon.png" // URL по умолчанию
		}

		users = append(users, user)
	}

	// Данные для передачи в шаблон
	data := PageData{
		Users: users,
	}

	if ok {
		// Получение данных текущего пользователя, если сессия есть
		var loggedUser LoggedUserStruct
		err = db.QueryRow("SELECT id, name, password, photos, description, userphoto FROM `all_users` WHERE id = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto)
		if err != nil {
			http.Error(w, "Ошибка получения данных текущего пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Преобразование ID фотографии текущего пользователя в URL
		var loggedUserphotoURL string
		if loggedUser.Userphoto != 0 {
			loggedUserphotoURL = "/userphoto/" + strconv.Itoa(int(loggedUser.Userphoto))
		} else {
			loggedUserphotoURL = "/static/icons/profile_page-icon.png"
		}

		data.LoggedUserId = int(currentUserID)
		data.LoggedUserphotoURL = loggedUserphotoURL
		data.UserphotoURL = loggedUserphotoURL
	}

	// Рендеринг шаблона с данными
	err = t.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
	}
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

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `all_users` (`name`, `password`, `photos`, `description`, `userphoto`) VALUES ('%s', '%s', '0', '%s', 0)", username, password, description))
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
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto)
	if err != nil {
		// http.Redirect(w, r, "/", http.StatusSeeOther)
		// return
	}

	photoIDs := strings.Split(user.Photos, ",")
	var photoURLs []string
	for _, id := range photoIDs {
		id = strings.TrimSpace(id)
		if id != "0" && id != "" {
			photoURLs = append(photoURLs, "/photo/"+id)
		}
	}

	var loggedUser LoggedUserStruct
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto)
	// if err != nil {
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }

	var loggedUserphotoURL string
	LoggedUserphotoID := loggedUser.Userphoto

	fmt.Println(strconv.Itoa(int(LoggedUserphotoID)))

	if LoggedUserphotoID != 0 {
		loggedUserphotoURL = "/userphoto/" + strconv.Itoa(int(LoggedUserphotoID))
	} else {
		loggedUserphotoURL = "/static/icons/profile_page-icon.png"
	}

	var userphotoURL string
	userphotoID := user.Userphoto

	fmt.Println(strconv.Itoa(int(userphotoID)))

	if userphotoID != 0 {
		userphotoURL = "/userphoto/" + strconv.Itoa(int(userphotoID))
	} else {
		userphotoURL = "/static/icons/profile_page-icon.png"
	}

	isOwner := loggedIn && currentUserID == user.Id

	data := struct {
		Profile
		IsOwner            bool
		LoggedUserId       int
		PhotoURLs          []string
		UserphotoURL       string
		LoggedUserphotoURL string
	}{
		Profile:            user,
		IsOwner:            isOwner,
		LoggedUserId:       loggedUserID,
		PhotoURLs:          photoURLs,
		UserphotoURL:       userphotoURL,
		LoggedUserphotoURL: loggedUserphotoURL,
	}

	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/secondHeader.html", "templates/footer.html")
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
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `name` = ? AND `password` = ?", username, password).Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.Photos, &dbUser.Description, &dbUser.Userphoto)
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
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var loggedUser LoggedUserStruct
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var loggedUserphotoURL string
	LoggedUserphotoID := loggedUser.Userphoto

	fmt.Println(strconv.Itoa(int(LoggedUserphotoID)))

	if LoggedUserphotoID != 0 {
		loggedUserphotoURL = "/userphoto/" + strconv.Itoa(int(LoggedUserphotoID))
	} else {
		loggedUserphotoURL = "/static/icons/profile_page-icon.png"
	}

	isOwner := loggedIn && currentUserID == user.Id

	data := struct {
		Profile
		IsOwner            bool
		LoggedUserId       int
		LoggedUserphotoURL string
	}{
		Profile:            user,
		IsOwner:            isOwner,
		LoggedUserId:       loggedUserID,
		LoggedUserphotoURL: loggedUserphotoURL,
	}

	t, err := template.ParseFiles("templates/settings.html", "templates/header.html", "templates/footer.html", "templates/secondHeader.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "settings", data)

}

func servePhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	photoID := vars["photoID"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var photoData []byte
	err = db.QueryRow("SELECT `photo` FROM `photos` WHERE `id` = ?", photoID).Scan(&photoData)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(photoData)
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

func create(w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)
	currentUserID := session.Values["user_id"].(uint16)

	loggedUserID := int(currentUserID)

	fmt.Print(currentUserID)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var loggedUser LoggedUserStruct
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var userphotoURL string
	userphotoID := loggedUser.Userphoto

	fmt.Println(strconv.Itoa(int(userphotoID)))

	if userphotoID != 0 {
		userphotoURL = "/userphoto/" + strconv.Itoa(int(userphotoID))
	} else {
		userphotoURL = "/static/icons/profile_page-icon.png"
	}
	user.UserphotoURL = userphotoURL

	var loggedUserphotoURL string
	LoggedUserphotoID := loggedUser.Userphoto

	fmt.Println(strconv.Itoa(int(LoggedUserphotoID)))

	if LoggedUserphotoID != 0 {
		loggedUserphotoURL = "/userphoto/" + strconv.Itoa(int(LoggedUserphotoID))
	} else {
		loggedUserphotoURL = "/static/icons/profile_page-icon.png"
	}

	data := struct {
		Profile
		LoggedUserId       int
		LoggedUserphotoURL string
	}{
		Profile:            user,
		LoggedUserId:       loggedUserID,
		LoggedUserphotoURL: loggedUserphotoURL,
	}

	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	fmt.Println(user)

	t.ExecuteTemplate(w, "create", data)
}

func createPhoto(w http.ResponseWriter, r *http.Request) {
	// Ограничение на размер загружаемого файла
	r.ParseMultipartForm(10 << 20) // 10 MB

	// Получаем файл и заголовки
	file, _, err := r.FormFile("photo")
	if err != nil {
		fmt.Fprintf(w, "Error retrieving the file")
		return
	}
	defer file.Close()

	// Чтение файла в буфер
	photoData, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(w, "Error reading the file")
		return
	}

	// Подключение к БД
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Вставка фотографии в базу данных
	stmt, err := db.Prepare("INSERT INTO `photos` (`photo`) VALUES (?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Выполняем запрос и получаем ID добавленной фотографии
	res, err := stmt.Exec(photoData)
	if err != nil {
		panic(err)
	}
	photoID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	// Получаем ID текущего пользователя из сессии
	session := GetSession(w, r)
	currentUserID := session.Values["user_id"].(uint16)

	// Получаем текущий список ID фотографий пользователя
	var currentPhotos string
	err = db.QueryRow("SELECT `photos` FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&currentPhotos)
	if err != nil {
		panic(err)
	}

	// Обновляем список фотографий
	if currentPhotos == "0" || currentPhotos == "" { // Если нет фото, заменяем "0" или пустое значение на ID новой фотографии
		currentPhotos = fmt.Sprintf("%d", photoID)
	} else {
		currentPhotos = fmt.Sprintf("%s,%d", currentPhotos, photoID)
	}

	// Обновляем запись в таблице all_users
	_, err = db.Exec("UPDATE `all_users` SET `photos` = ? WHERE `id` = ?", currentPhotos, currentUserID)
	if err != nil {
		panic(err)
	}

	// Перенаправление на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addUserphoto(w http.ResponseWriter, r *http.Request) {
	// Ограничение на размер загружаемого файла
	r.ParseMultipartForm(10 << 20) // 10 MB

	// Получаем файл и заголовки
	file, _, err := r.FormFile("photo")
	if err != nil {
		fmt.Fprintf(w, "Error retrieving the file")
		return
	}
	defer file.Close()

	// Чтение файла в буфер
	photoData, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(w, "Error reading the file")
		return
	}

	// Подключение к БД
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Вставка фотографии в базу данных
	stmt, err := db.Prepare("INSERT INTO `userphotos` (`photo`) VALUES (?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Выполняем запрос и получаем ID добавленной фотографии
	res, err := stmt.Exec(photoData)
	if err != nil {
		panic(err)
	}
	photoID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	// Получаем ID текущего пользователя из сессии
	session := GetSession(w, r)
	currentUserID := session.Values["user_id"].(uint16)

	// Получаем текущий список ID фотографий пользователя
	// var currentPhotos uint16
	// err = db.QueryRow("SELECT `userphoto` FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&currentPhotos)
	// if err != nil {
	// 	panic(err)
	// }

	// Обновляем запись в таблице all_users
	_, err = db.Exec("UPDATE `all_users` SET `userphoto` = ? WHERE `id` = ?", photoID, currentUserID)
	if err != nil {
		panic(err)
	}

	// Перенаправление на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func serveUserphoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	photoID := vars["photoID"]

	// fmt.Println(photoID)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var photoData []byte
	err = db.QueryRow("SELECT `photo` FROM `userphotos` WHERE `id` = ?", photoID).Scan(&photoData)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(photoData)
}

func deleteAccount(w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)
	currentUserID, ok := session.Values["user_id"].(uint16)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
		return
	}

	fmt.Println("Текущий пользователь ID:", currentUserID)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/golang")
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM all_users WHERE id = ?")
	if err != nil {
		http.Error(w, "Ошибка при подготовке запроса", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(currentUserID)
	if err != nil {
		fmt.Println("Ошибка при удалении пользователя из базы данных:", err)
		http.Error(w, "Ошибка удаления аккаунта", http.StatusInternalServerError)
		return
	}

	ClearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleFunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/register", register).Methods("GET")
	rtr.HandleFunc("/login", login).Methods("GET", "POST")
	rtr.HandleFunc("/logout", logout).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/createPhoto", createPhoto).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/addUserphoto", addUserphoto).Methods("POST")
	rtr.HandleFunc("/reg_user", reg_user).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/change_desc", change_desc).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}", user_profile).Methods("GET")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/settings", user_settings).Methods("GET")
	rtr.HandleFunc("/log_user", logUser).Methods("GET") // Новый маршрут
	rtr.HandleFunc("/photo/{photoID:[0-9]+}", servePhoto).Methods("GET")
	rtr.HandleFunc("/userphoto/{photoID:[0-9]+}", serveUserphoto).Methods("GET")
	rtr.HandleFunc("/deleteaccount", deleteAccount).Methods("GET")

	http.Handle("/", rtr)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.ListenAndServe(":8080", nil)
}

func main() {
	handleFunc()
}
