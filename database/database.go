package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"log"
)

type User struct {
	Name       string
	Username   string
	Email      string
	Password   string
	Errors     map[string]string
}

type Token struct {
	Token	string
	Expires string
}

func InitializeDatabase() {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		internal_id int NOT NULL AUTO_INCREMENT,
		name varchar(50),
		username varchar(15) NOT NULL UNIQUE,
		email varchar(255),
		password varchar(60),
		PRIMARY KEY (internal_id)
		);`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		internal_id int NOT NULL AUTO_INCREMENT,
		token blob,
		username varchar(15) NOT NULL UNIQUE,
		expires timestamp,
		PRIMARY KEY (internal_id)
		);`)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertToken(w http.ResponseWriter, r *http.Request, u User, tk Token) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	TokenInsert, err := db.Prepare(`
	INSERT INTO tokens (username, token, expires) VALUES ( ?, ?, ? ) ON DUPLICATE KEY UPDATE token = ?, expires = ?
	`)

	_, err = TokenInsert.Exec(u.Username, tk.Token, tk.Expires, tk.Token, tk.Expires)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request, u User) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	UserInsert, err := db.Prepare(`
	INSERT INTO users (name, username, email, password) VALUES ( ?, ?, ?, ? )
	`)

	_, err = UserInsert.Exec(u.Name, u.Username, u.Email, u.Password)
	if err != nil {
		http.Redirect(w, r, "/users/create/", http.StatusFound)
	}
	http.Redirect(w, r, "/users/login/", http.StatusFound)
}

func GetSessionKey(w http.ResponseWriter, r *http.Request, token string) string {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var username string

	err = db.QueryRow(`
	SELECT username FROM tokens WHERE token = ?
	`, token).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
	}
	return username
}

func GetHashedPwdForUser(w http.ResponseWriter, r *http.Request, u User) string {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var hashedPwd string

	err = db.QueryRow(`
	SELECT password FROM users WHERE username = ?
	`, u.Username).Scan(&hashedPwd)
	if err != nil {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
	}
	return hashedPwd
}