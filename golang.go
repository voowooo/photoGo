package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type Profile struct {
	Id, Userphoto                                                     uint16
	Username, Password, Photos, Description, Subscriptions, Followers string
	UserphotoURL                                                      string
	Color                                                             string
}

type CommentStruct struct {
	Id    uint16
	Text  string
	Owner string
}

type LoggedUserStruct struct {
	Id, Userphoto                           uint16
	Username, Password, Photos, Description string
	UserphotoURL                            string
	Color                                   string
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

func GetFullInfoAboutLoggedUser(w http.ResponseWriter, r *http.Request) (Profile, []string, []string, error) {
	// Подключение к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		return Profile{}, nil, nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}
	defer db.Close()

	// Получение ID текущего пользователя из сессии
	session := GetSession(w, r) // Передаем оба аргумента: w и r
	currentUserID, ok := session.Values["user_id"].(uint16)
	if !ok {
		return Profile{}, nil, nil, fmt.Errorf("не удалось получить ID пользователя из сессии")
	}

	// Запрос данных пользователя
	var loggedUser Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).
		Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto, &loggedUser.Subscriptions, &loggedUser.Followers, &loggedUser.Color)
	if err != nil {
		return Profile{}, nil, nil, fmt.Errorf("ошибка при выполнении запроса к базе данных: %w", err)
	}

	if loggedUser.Userphoto != 0 {
		loggedUser.UserphotoURL = "/userphoto/" + strconv.Itoa(int(loggedUser.Userphoto))
	} else {
		loggedUser.UserphotoURL = "/static/icons/profile_page-icon.png" // URL по умолчанию
	}

	names := idToNames(loggedUser.Subscriptions)
	followers := idToNames(loggedUser.Followers)

	return loggedUser, names, followers, nil
}

// Основной обработчик, который рендерит шаблон
// Основной обработчик, который рендерит шаблон
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("index")

	// Парсинг шаблонов

	// Получение данных о текущем пользователе
	loggedUser, subs, followers, err := GetFullInfoAboutLoggedUser(w, r)
	if err != nil {
		// http.Error(w, "Ошибка при получении информации о пользователе: "+err.Error(), http.StatusInternalServerError)
		http.Redirect(w, r, "/allusers", http.StatusSeeOther)
		return
	}

	session := GetSession(w, r)
	currentUserID, loggedIn := session.Values["user_id"].(uint16)

	isOwner := loggedIn && currentUserID == loggedUser.Id

	data := struct {
		Profile
		LoggedUserId       int
		LoggedUserphotoURL string
		IsOwner            bool
		Subs               []string
		Followers          []string
	}{
		Profile:            loggedUser,
		LoggedUserId:       int(loggedUser.Id),
		LoggedUserphotoURL: loggedUser.UserphotoURL,
		IsOwner:            isOwner,
		Subs:               subs,
		Followers:          followers,
	}

	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/secondHeader.html")
	if err != nil {
		http.Error(w, "Ошибка при загрузке шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Передача данных в шаблон
	err = t.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func allusers(w http.ResponseWriter, r *http.Request) {
	// Парсинг HTML-шаблонов
	t, err := template.ParseFiles("templates/allusers.html", "templates/header.html", "templates/footer.html", "templates/secondHeader.html")
	// t, err := template.ParseFiles("templatesNEW/index.html", "templatesNEW/header.html", "templatesNEW/footer.html")
	if err != nil {
		http.Error(w, "Ошибка при загрузке шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение сессии
	session := GetSession(w, r)
	currentUserID, ok := session.Values["user_id"].(uint16)

	// Подключение к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Запрос на получение всех пользователей из таблицы `all_users`
	rows, err := db.Query("SELECT id, name, password, photos, description, userphoto, color FROM `all_users`")
	if err != nil {
		http.Error(w, "Ошибка выполнения запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание пустого массива пользователей
	var users []Profile
	for rows.Next() {
		var user Profile
		err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto, &user.Color)
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
	err = t.ExecuteTemplate(w, "allusers", data)
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
	fmt.Println("log")
	if r.Method == http.MethodPost {
		fmt.Println("logff")
		username := r.FormValue("username")
		password := r.FormValue("password")

		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

		fmt.Println("log")

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
	fmt.Println("logout")
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `all_users` (`name`, `password`, `photos`, `description`, `userphoto`, `subscriptions`, `followers`, `color`) VALUES ('%s', '%s', '0', '%s', 0, '', '', '')", username, password, description))
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

	fmt.Println(loggedUserID)

	vars := mux.Vars(r)
	requestedUserID := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto, &user.Subscriptions, &user.Followers, &user.Color)
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

	// var loggedUser LoggedUserStruct
	var loggedUser Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto, &loggedUser.Subscriptions, &loggedUser.Followers, &loggedUser.Color)
	// if err != nil {
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }

	isIfollow := isIfollowCheck(loggedUserID, requestedUserID)

	names := idToNames(user.Subscriptions)

	followers := idToNames(user.Followers)

	fmt.Println("isIfollow")
	fmt.Println(isIfollow)

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
		Names              []string
		Followers          []string
		IsIFollow          int
	}{
		Profile:            user,
		IsOwner:            isOwner,
		LoggedUserId:       loggedUserID,
		PhotoURLs:          photoURLs,
		UserphotoURL:       userphotoURL,
		LoggedUserphotoURL: loggedUserphotoURL,
		Names:              names,
		Followers:          followers,
		IsIFollow:          isIfollow,
	}

	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/secondHeader.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "profile", data)
}

func isIfollowCheck(me int, req string) int {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
	}
	defer db.Close() // Закрываем соединение по завершении функции

	var mySubs string
	err = db.QueryRow("SELECT `subscriptions` FROM `all_users` WHERE `id` = ?", me).Scan(&mySubs)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("error")
		} else {
			fmt.Println("Ошибка запроса:", err)
		}
	}

	parts := strings.Split(mySubs, ",")

	found := false
	for _, part := range parts {
		if part == req {
			found = true
			break
		}
	}

	if found {
		return 1
	} else {
		return 0
	}

}

func idToNames(ids string) []string {
	// Разделяем входную строку на отдельные строки ID
	parts := strings.Split(ids, ",")

	// Создаем срез для числовых ID
	numbers := make([]int, len(parts))

	// Преобразуем каждую часть в число
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			fmt.Println("Ошибка преобразования:", err)
			return nil
		}
		numbers[i] = num
	}

	// Подключаемся к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return nil
	}
	defer db.Close() // Закрываем соединение по завершении функции

	// Создаем срез для хранения имен и фотографий
	results := make([]string, 0, len(numbers))

	// Выполняем запрос для каждого ID и добавляем имя и фото в срез
	for _, id := range numbers {
		var name string
		var userphotoID int
		err = db.QueryRow("SELECT `name`, `userphoto` FROM `all_users` WHERE `id` = ?", id).Scan(&name, &userphotoID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("ID %d не найден в базе данных.\n", id)
			} else {
				fmt.Println("Ошибка запроса:", err)
			}
			continue
		}

		// Получаем данные фотографии из таблицы `userphotos`
		var photoData []byte
		err = db.QueryRow("SELECT `photo` FROM `userphotos` WHERE `id` = ?", userphotoID).Scan(&photoData)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("Фото с ID %d не найдено в таблице userphotos.\n", userphotoID)
			} else {
				fmt.Println("Ошибка запроса фотографии:", err)
			}
			continue
		}

		// Конвертируем фото в base64 строку
		photoBase64 := base64.StdEncoding.EncodeToString(photoData)
		// Создаем строку для тега <img> с base64-кодированной фотографией
		photoHTML := "data:image/jpeg;base64," + photoBase64

		var subColor string
		err = db.QueryRow("SELECT `color` FROM `all_users` WHERE `id` = ?", id).Scan(&subColor)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("цвет подписчика не найден")
			} else {
				fmt.Println("Ошибка запроса цвета:", err)
			}
			continue
		}

		var SubPhotos string
		err = db.QueryRow("SELECT `photos` FROM `all_users` WHERE `id` = ?", id).Scan(&SubPhotos)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("фото %d не найдены в базе данных.\n", id)
			} else {
				fmt.Println("Ошибка запроса:", err)
			}
			continue
		}

		// Разделяем `SubPhotos`, чтобы получить список ID фотографий
		SubPhotosIds := strings.Split(SubPhotos, ",")

		photoLink := make([]string, 0, len(SubPhotosIds))

		// Если фото три или меньше, добавляем их все, пропуская нулевые
		if len(SubPhotosIds) <= 3 {
			for _, photoID := range SubPhotosIds {
				if photoID != "0" && photoID != "" { // Пропускаем нулевые и пустые ID
					// photoLink = append(photoLink, "<img src='http://192.168.56.214:8080/photo/"+photoID+"' class='index-sub-last-photos'>")
					photoLink = append(photoLink, "<img src='/photo/"+photoID+"' class='index-sub-last-photos'>")
				}
			}
		} else {
			// Если больше трёх фото, выбираем последние три, пропуская нулевые
			SubPhotosIds = SubPhotosIds[len(SubPhotosIds)-3:]
			for _, photoID := range SubPhotosIds {
				if photoID != "0" && photoID != "" { // Пропускаем нулевые и пустые ID
					photoLink = append(photoLink, "<img src='/photo/"+photoID+"' class='index-sub-last-photos'>")
				}
			}
		}

		var photosHTML string
		if len(photoLink) == 0 {
			photosHTML = "<p>Нет фото</p>" // Сообщение, если фото нет
		} else {
			photosHTML = strings.Join(photoLink, "") // Объединяем фото без разделителя
		}

		// Создаем HTML-код с ссылкой на профиль и фото
		profileHTML := "<div>" +
			"<a href='/profile/" + strconv.Itoa(id) + "' class='profile-sub' style='background-color: rgb(" + subColor + ");'>" +
			"<img src='" + photoHTML + "' alt='" + name + "' class='user-photo index-userphoto'>" +
			"<h2>" + name + "</h2>" +
			"</a>" +
			"<div class='index-prev-photos' style='background-color: rgb(" + subColor + ");'>" +
			photosHTML +
			"</div>" +
			"</div>"

		// Добавляем результат в срез
		results = append(results, profileHTML)
	}

	// Выводим и возвращаем массив результатов
	// fmt.Println(results)
	return results
}

func logUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var dbUser Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `name` = ? AND `password` = ?", username, password).Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.Photos, &dbUser.Description, &dbUser.Userphoto, &dbUser.Subscriptions, &dbUser.Followers, &dbUser.Color)
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", requestedUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto, &user.Subscriptions, &user.Followers, &user.Color)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var loggedUser Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto, &loggedUser.Subscriptions, &loggedUser.Followers, &loggedUser.Color)
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

func change_color(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	Color := r.FormValue("pickedColor")
	Id := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Use a prepared statement to prevent SQL injection
	stmt, err := db.Prepare("UPDATE `all_users` SET `color` = ? WHERE `id` = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Execute the query with user-provided values
	_, err = stmt.Exec(Color, Id)
	if err != nil {
		panic(err)
	}

	// Redirect to the user's profile or wherever appropriate
	http.Redirect(w, r, fmt.Sprintf("/profile/%s", Id), http.StatusSeeOther)
}

func change_desc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	description := r.FormValue("description")
	Id := vars["user_id"]

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var user Profile
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&user.Id, &user.Username, &user.Password, &user.Photos, &user.Description, &user.Userphoto, &user.Subscriptions, &user.Followers, &user.Color)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var loggedUser LoggedUserStruct
	err = db.QueryRow("SELECT * FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&loggedUser.Id, &loggedUser.Username, &loggedUser.Password, &loggedUser.Photos, &loggedUser.Description, &loggedUser.Userphoto, &user.Subscriptions, &user.Followers, &loggedUser.Color)
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
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Вставка фотографии в базу данных
	stmt, err := db.Prepare("INSERT INTO `photos` (`photo`, `comments`) VALUES (?, ?)")
	// stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO `comments` (`text`, `owner`) VALUES ('%s', '%d')", content, owner))
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Выполняем запрос и получаем ID добавленной фотографии
	res, err := stmt.Exec(photoData, "joi")
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
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
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

func addComment(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("commContent")
	PhotoId := r.FormValue("PhotoId")
	ProfileId := r.FormValue("ProfileId")

	session := GetSession(w, r)
	owner := int(session.Values["user_id"].(uint16))

	fmt.Println("newComment")
	fmt.Println(content)
	fmt.Println(owner)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// insert, err := db.Query(fmt.Sprintf("INSERT INTO `comments` (`text`, `owner`) VALUES ('%s', '%d')", content, owner))
	// if err != nil {
	// 	panic(err)
	// }
	// defer insert.Close()

	stmt, err := db.Prepare("INSERT INTO `comments` (`text`, `owner`) VALUES (?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(content, owner)
	if err != nil {
		panic(err)
	}
	CommentId, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	fmt.Println("commentId")
	fmt.Println(CommentId)

	fmt.Println(PhotoId)

	// Получаем текущий список ID фотографий пользователя
	var currentComments string
	err = db.QueryRow("SELECT `comments` FROM `photos` WHERE `id` = ?", PhotoId).Scan(&currentComments)
	if err != nil {
		panic(err)
	}

	// Обновляем список фотографий
	if currentComments == "0" || currentComments == "" { // Если нет фото, заменяем "0" или пустое значение на ID новой фотографии
		currentComments = fmt.Sprintf("%d", CommentId)
	} else {
		currentComments = fmt.Sprintf("%s,%d", currentComments, CommentId)
	}

	// Обновляем запись в таблице all_users
	_, err = db.Exec("UPDATE `photos` SET `comments` = ? WHERE `id` = ?", currentComments, PhotoId)
	if err != nil {
		panic(err)
	}

	RedirectUrl := "/profile/" + ProfileId
	http.Redirect(w, r, RedirectUrl, http.StatusSeeOther)

}

func fullPhoto(w http.ResponseWriter, r *http.Request) {
	fmt.Println("full")

	// Получаем текущий путь URL
	urlPath := r.URL.Path
	fmt.Println("URL Path:", urlPath)

	// Разделяем строку по символу '/'
	parts := strings.Split(urlPath, "/")

	photoID := ""

	// Проверяем, что в массиве достаточно частей, и получаем последнюю часть (это должно быть число)
	if len(parts) > 0 {
		photoID = parts[len(parts)-1]
		fmt.Println("Photo ID:", photoID)
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var CurrentPhotoCommentsIds string
	err = db.QueryRow("SELECT comments FROM photos WHERE id = ?", photoID).Scan(&CurrentPhotoCommentsIds)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Println("Comments IDs:", CurrentPhotoCommentsIds)

	// Преобразуем строку с идентификаторами комментариев в массив
	ids := strings.Split(CurrentPhotoCommentsIds, ",")
	if len(ids) == 0 {
		fmt.Println("Нет комментариев для этой фотографии")
		w.WriteHeader(http.StatusOK)                 // Возвращаем пустой массив, если комментариев нет
		json.NewEncoder(w).Encode([]CommentStruct{}) // Возвращаем пустой JSON-массив
		return
	}

	// Создаем параметры для IN-оператора
	query, args := createInQuery("SELECT id, text, owner FROM comments WHERE id IN (%s)", ids)
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Чтение всех комментариев
	comments := []CommentStruct{}
	for rows.Next() {
		var comment CommentStruct
		err := rows.Scan(&comment.Id, &comment.Text, &comment.Owner)
		if err != nil {
			http.Error(w, "Ошибка при чтении комментариев", http.StatusInternalServerError)
			return
		}

		// Выводим значение Owner перед конвертацией
		fmt.Printf("Owner перед конвертацией: %s\n", comment.Owner)

		comments = append(comments, comment)
	}

	// Теперь, когда мы прочитали все комментарии, обрабатываем их
	for i := range comments {
		// Пробуем конвертировать в целое число
		ownerId, err := strconv.Atoi(comments[i].Owner)
		if err != nil {
			fmt.Printf("Ошибка конвертации Owner в число: %v, значение Owner: %s\n", err, comments[i].Owner)
			http.Error(w, "Ошибка при конвертации Owner в число", http.StatusInternalServerError)
			return
		}

		fmt.Println("ownerId:", ownerId)

		var username string
		errr := db.QueryRow("SELECT name FROM all_users WHERE id = ?", ownerId).Scan(&username)
		if errr != nil {
			if errr == sql.ErrNoRows {
				// Если не найдено ни одной строки
				fmt.Println("Пользователь не найден")
				username = "Неизвестный пользователь" // Обрабатываем неизвестных пользователей корректно
			} else {
				http.Error(w, "Ошибка при выполнении запроса", http.StatusInternalServerError)
				fmt.Println("error")
				return
			}
		}

		fmt.Println("Имя пользователя:", username)
		comments[i].Owner = username
		fmt.Printf("Comment ID: %d, Text: %s, Owner: %s\n", comments[i].Id, comments[i].Text, comments[i].Owner)
	}

	// Установите заголовок и код состояния после завершения всех обработок
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем комментарии в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(comments)
}

// Функция для создания SQL-запроса с IN-оператором
func createInQuery(baseQuery string, ids []string) (string, []interface{}) {
	query := fmt.Sprintf(baseQuery, strings.Repeat("?,", len(ids)-1)+"?")
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return query, args
}

func new_sub(w http.ResponseWriter, r *http.Request) {
	fmt.Println("oke")

	session := GetSession(w, r)
	currentUserID := session.Values["user_id"].(uint16)

	fmt.Println(currentUserID)

	vars := mux.Vars(r)
	userIDStr := vars["user_id"]

	// Преобразуем user_id из строки в int
	userSubTo, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат числа", http.StatusBadRequest)
		return
	}

	fmt.Printf("User ID from URL: %d\n", userSubTo)

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/photogo")
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var currentSubs string
	err = db.QueryRow("SELECT `subscriptions` FROM `all_users` WHERE `id` = ?", currentUserID).Scan(&currentSubs)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Обновляем список фотографий
	if currentSubs == "0" || currentSubs == "" { // Если нет фото, заменяем "0" или пустое значение на ID новой фотографии
		currentSubs = fmt.Sprintf("%d", userSubTo)
	} else {
		currentSubs = fmt.Sprintf("%s,%d", currentSubs, userSubTo)
	}

	// Обновляем запись в таблице all_users
	_, err = db.Exec("UPDATE `all_users` SET `subscriptions` = ? WHERE `id` = ?", currentSubs, currentUserID)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//add follower
	var userSubToFollowers string
	err = db.QueryRow("SELECT `followers` FROM `all_users` WHERE `id` = ?", userSubTo).Scan(&userSubToFollowers)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Обновляем список фотографий
	if userSubToFollowers == "0" || userSubToFollowers == "" { // Если нет фото, заменяем "0" или пустое значение на ID новой фотографии
		userSubToFollowers = fmt.Sprintf("%d", currentUserID)
	} else {
		userSubToFollowers = fmt.Sprintf("%s,%d", userSubToFollowers, currentUserID)
	}

	// Обновляем запись в таблице all_users
	_, err = db.Exec("UPDATE `all_users` SET `followers` = ? WHERE `id` = ?", userSubToFollowers, userSubTo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	RedirectUrl := "/profile/" + userIDStr
	http.Redirect(w, r, RedirectUrl, http.StatusSeeOther)

}

func handleFunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/allusers", allusers).Methods("GET")
	rtr.HandleFunc("/register", register).Methods("GET")
	rtr.HandleFunc("/login", login).Methods("GET", "POST")
	rtr.HandleFunc("/logout", logout).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/createPhoto", createPhoto).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/addUserphoto", addUserphoto).Methods("POST")
	rtr.HandleFunc("/reg_user", reg_user).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/change_desc", change_desc).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/changeUserColor", change_color).Methods("POST")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}", user_profile).Methods("GET")
	rtr.HandleFunc("/profile/{user_id:[0-9]+}/settings", user_settings).Methods("GET")
	rtr.HandleFunc("/sub/{user_id:[0-9]+}", new_sub).Methods("POST")
	rtr.HandleFunc("/log_user", logUser).Methods("GET") // Новый маршрут
	rtr.HandleFunc("/photo/{photoID:[0-9]+}", servePhoto).Methods("GET")
	rtr.HandleFunc("/userphoto/{photoID:[0-9]+}", serveUserphoto).Methods("GET")
	rtr.HandleFunc("/deleteaccount", deleteAccount).Methods("GET")
	rtr.HandleFunc("/profile/addComment", addComment).Methods("POST")
	rtr.HandleFunc("/fullphoto/photo/{photoID:[0-9]+}", fullPhoto).Methods("GET")

	http.Handle("/", rtr)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println(port)

	// http.ListenAndServe("192.168.56.214:8080", nil)
	// http.ListenAndServe("134.17.129.247:" + port, nil)

	http.ListenAndServe(":"+port, nil)
}

func main() {
	handleFunc()
}
