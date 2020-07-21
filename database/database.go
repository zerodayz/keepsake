package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	Name         string
	Username     string
	Email        string
	Password     string
	UserLoggedIn string
	Errors       map[string]string
}

type Token struct {
	Token   string
	Expires string
}

type WikiPage struct {
	InternalId     int
	Title          string
	Content        string
	Username       string
	DateCreated    string
	LastModified   string
	LastModifiedBy string
	Deleted        int
}

type WikiPageRevision struct {
	InternalId     int
	WikiPageId	   int
	RevisionId     int
	DateModified   string
	Title          string
	Content        string
	Username       string
	DateCreated    string
	LastModified   string
	LastModifiedBy string
	Deleted        int
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
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS pages (
		internal_id int NOT NULL AUTO_INCREMENT,
		title varchar(50) NOT NULL,
		content TEXT,
		username varchar(15) NOT NULL,
		deleted int,
		last_modified_by varchar(15),
		last_modified timestamp,
		date_created timestamp,
		PRIMARY KEY (internal_id)
		);`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS pages_rev (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int,
		revision_id int,
		date_modified timestamp,
		title varchar(50) NOT NULL,
		content TEXT,
		username varchar(15) NOT NULL,
		deleted int,
		last_modified_by varchar(15),
		last_modified timestamp,
		date_created timestamp,
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

func CreatePage(w http.ResponseWriter, r *http.Request, s WikiPage) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Set deleted to 0 during creation.
	s.Deleted = 0

	PageInsert, err := db.Prepare(`
	INSERT INTO pages (title, content, username, deleted, date_created) VALUES ( ?, ?, ?, ?, ? )
	`)

	var res sql.Result

	res, err = PageInsert.Exec(s.Title, s.Content, s.Username, s.Deleted, s.DateCreated)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	wikiPageId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	PageRevisionInsert, err := db.Prepare(`
	INSERT INTO pages_rev (wiki_page_id, revision_id, title, content, username, deleted, date_created)
	VALUES ( ?, ?, ?, ?, ?, ?, ? )
	`)

	_, err = PageRevisionInsert.Exec(wikiPageId, 1, s.Title, s.Content, s.Username, s.Deleted, s.DateCreated)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
}

func UpdatePage(w http.ResponseWriter, r *http.Request, s WikiPage) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var title, content, username, deleted, lastModified, dateCreated, revisionId string
	// Get the original page
	err = db.QueryRow(`
	SELECT title, content, username, deleted, last_modified, date_created
	FROM pages WHERE internal_id = ?`, s.InternalId).Scan(&title, &content, &username, &deleted,
		&lastModified, &dateCreated)
	if err != nil {
		log.Fatal(err)
	}

	// Get latest revision_number
	rows, err := db.Query("SELECT revision_id FROM pages_rev WHERE wiki_page_id = ?", s.InternalId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&revisionId)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	i, err := strconv.Atoi(revisionId)
	i++

	// Insert into revisions
	PageRevisionInsert, err := db.Prepare(`
	INSERT INTO pages_rev (wiki_page_id, revision_id, title, content, username, deleted, date_created, last_modified_by, last_modified)
	VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )
	`)

	_, err = PageRevisionInsert.Exec(s.InternalId, i, s.Title, s.Content, username, deleted, dateCreated, s.LastModifiedBy, s.LastModified)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	// Set deleted to 0 for newly updated.
	s.Deleted = 0

	PageUpdate, err := db.Prepare(`
	UPDATE pages SET title = ?, content = ?, deleted = ?, last_modified = ?, last_modified_by = ?
	WHERE internal_id = ?
	`)

	_, err = PageUpdate.Exec(s.Title, s.Content, s.Deleted, s.LastModified, s.LastModifiedBy, s.InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func DeletePage(w http.ResponseWriter, r *http.Request, InternalId int) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	PageUpdate, err := db.Prepare(`
	UPDATE pages SET deleted = ?
	WHERE internal_id = ?
	`)

	_, err = PageUpdate.Exec(1, InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func ShowRevisionPage(w http.ResponseWriter, r *http.Request, InternalId int) *WikiPageRevision {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var title, content, dateCreated, lastModified, lastModifiedBy, username string

	err = db.QueryRow(`
	SELECT title FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&title)
	err = db.QueryRow(`
	SELECT content FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&content)
	err = db.QueryRow(`
	SELECT date_created FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&dateCreated)
	err = db.QueryRow(`
	SELECT last_modified FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&lastModified)
	err = db.QueryRow(`
	SELECT last_modified_by FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&lastModifiedBy)
	err = db.QueryRow(`
	SELECT username FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	return &WikiPageRevision{Title: title, Content: content, DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, Username: username}
}

func ShowPage(w http.ResponseWriter, r *http.Request, InternalId int) *WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var title, content, dateCreated, lastModified, lastModifiedBy, username string

	err = db.QueryRow(`
	SELECT title FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&title)
	err = db.QueryRow(`
	SELECT content FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&content)
	err = db.QueryRow(`
	SELECT date_created FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&dateCreated)
	err = db.QueryRow(`
	SELECT last_modified FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&lastModified)
	err = db.QueryRow(`
	SELECT last_modified_by FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&lastModifiedBy)
	err = db.QueryRow(`
	SELECT username FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	return &WikiPage{Title: title, Content: content, DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, Username: username}
}

func FetchDeletedPages(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages []WikiPage
		id int
		title string
	)
	rows, err := db.Query("SELECT internal_id, title FROM pages WHERE deleted = ?", 1)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Title: title})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func FetchRevisionPages(w http.ResponseWriter, r *http.Request, internalId int) ([]WikiPageRevision, string) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var wikiPage string

	err = db.QueryRow(`
	SELECT title FROM pages WHERE internal_id = ?
	`, internalId).Scan(&wikiPage)

	var (
		wikiPages []WikiPageRevision
		revisionId int
		title string
		dateModified string
		lastModifiedBy string
	)
	rows, err := db.Query(`SELECT internal_id, revision_id, title, date_modified, COALESCE(last_modified_by, '') as last_modified_by
		FROM pages_rev WHERE wiki_page_id = ?`, internalId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&internalId, &revisionId, &title, &dateModified, &lastModifiedBy )
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPageRevision{RevisionId: revisionId, InternalId: internalId, Title: title, LastModifiedBy: lastModifiedBy, LastModified: dateModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages, wikiPage
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
