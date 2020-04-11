package users

import (
	"html/template"
	"net/http"
)

type User struct {
	InternalId string
	Name       string
	Username   string
	Email      string
	Password   string
	Errors     map[string]string
}

var (
	tmplpath  = "tmpl/"
	templates = template.Must(template.ParseFiles(tmplpath + "login.html"))
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login")
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.Execute(w, tmpl+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
