package users

import (
	"encoding/base64"
	"github.com/zerodayz/keepsake/database"
	"github.com/zerodayz/keepsake/pages"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var (
	tmplpath = "tmpl/users/"
)

func GenerateToken(email string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(hash)
}

func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// handlers

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	u := database.User{}

	username := pages.ReadCookie(w, r)
	if username != "Unauthorized" {
		p.UserLoggedIn, u.UserLoggedIn = username, username
	} else {
		p.UserLoggedIn = username
	}

	if r.Method == "POST" {
		r.ParseForm()
		u.Name = r.PostFormValue("name")
		u.Username = r.PostFormValue("username")
		u.Email = r.PostFormValue("email")
		u.Password = HashAndSalt([]byte(r.PostFormValue("password")))
		database.CreateUser(w, r, u)
	}

	t := template.Must(template.ParseFiles(tmplpath + "create.html"))
	err := t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	u := database.User{}
	tk := database.Token{}

	if r.Method == "POST" {
		r.ParseForm()
		u.Username = r.PostFormValue("username")
		u.Email = r.PostFormValue("email")
		u.Password = r.PostFormValue("password")
		hashedPwd := database.GetHashedPwdForUser(w, r, u)
		if ComparePasswords(hashedPwd, []byte(u.Password)) {

			uuid := GenerateToken(u.Email)
			expire := time.Now().AddDate(0, 0, 1).UTC()

			tk.Expires = expire.Format("20060102150405")
			tk.Token = uuid

			database.InsertToken(w, r, u, tk)

			cookie := http.Cookie{Name: "gowiki_session", Value: uuid, Path: "/", Expires: expire, MaxAge: 86400}
			http.SetCookie(w, &cookie)

			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	t := template.Must(template.ParseFiles(tmplpath + "login.html"))
	err := t.Execute(w, "login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{Name: "gowiki_session", Path: "/", MaxAge: -1}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/users/login/", http.StatusFound)
	return
}
