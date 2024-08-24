package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type User struct {
	Name                  string
	Age                   uint16
	Money                 int16
	Avg_grades, Happiness float64
	Hobbies               []string
}

func (u *User) getAllInfo() string {
	return fmt.Sprintf("User name is: %s. He is %d. And he has %d money", u.Name, u.Age, u.Money)
}

func (u *User) setNewName(newName string) {
	u.Name = newName
}

func home_page(w http.ResponseWriter, r *http.Request) {
	bob := User{Name: "Bob", Age: 25, Money: -50, Avg_grades: 4.2, Happiness: 0.8, Hobbies: []string{"Football", "Skate", "Dance"}}
	// bob.setNewName("Alex")
	// // fmt.Fprintf(w, bob.getAllInfo())
	// fmt.Fprintf(w, `<h1>Main Text<h1>`)
	// fmt.Fprintf(w, `<h1>Main yesy<h1>`)
	tmpl, _ := template.ParseFiles("templates/home_page.html")
	tmpl.Execute(w, bob)
}

func create_page(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create Page")
}

func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/create/", create_page)
	http.ListenAndServe(":8080", nil)
}

func main() {

	handleRequest()
}
