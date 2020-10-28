package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type User struct {
	Name         string
	Username     string
	Email        string
	Password     string
	UserLoggedIn string
	Errors       map[string]string
}

type File struct {
	Name string
}

type Token struct {
	Token   string
	Expires string
}

type Tag struct {
	InternalId  int
	Name        string
	DateCreated string
	CreatedBy   string
}

type Comment struct {
	InternalId  int
	WikiPageId  int
	CreatedBy   string
	DateCreated string
	Title       string
	Body        string
}

type WikiPage struct {
	InternalId     int
	WikiPageId     int
	CommentCount   int
	Liked		   int
	Repair		   int
	Title          string
	Content        string
	Tags           []string
	Username       string
	DateCreated    string
	LastModified   string
	LastModifiedBy string
	Deleted        int
	UserLoggedIn   string
	CreatedBy      string
	Body           string
	DisplayToday   template.HTML
	DisplayUpVoted template.HTML
	DisplayBody    template.HTML
	DisplayComment template.HTML
	Errors         map[string]string
}

type WikiPageRevision struct {
	InternalId     int
	WikiPageId     int
	RevisionId     int
	DateModified   string
	Title          string
	Content        string
	Tags           []string
	Username       string
	DateCreated    string
	LastModified   string
	LastModifiedBy string
	Deleted        int
	UserLoggedIn   string
	CreatedBy      string
	Body           string
	DisplayBody    template.HTML
	Errors         map[string]string
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
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int,
		created_by varchar(15) NOT NULL,
		date_created timestamp,
		title varchar(255) NOT NULL,
		content TEXT,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS likes (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int,
		username varchar(15) NOT NULL,
		status int,
		PRIMARY KEY (internal_id),
		CONSTRAINT wiki_id_user UNIQUE (username , wiki_page_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS repairs (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int NOT NULL UNIQUE,
		username varchar(15) NOT NULL,
		status int,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
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
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tags (
		internal_id int NOT NULL AUTO_INCREMENT,
		name varchar(60) NOT NULL UNIQUE,
		date_created timestamp,
		created_by varchar(15) NOT NULL,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS pages (
		internal_id int NOT NULL AUTO_INCREMENT,
		title varchar(255) NOT NULL,
		content TEXT,
		tags TEXT,
		created_by varchar(15) NOT NULL,
		deleted int,
		last_modified_by varchar(15),
		last_modified timestamp,
		date_created timestamp,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS pages_preview (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int,
		title varchar(255) NOT NULL,
		content TEXT,
		tags TEXT,
		created_by varchar(15) NOT NULL,
		deleted int,
		last_modified_by varchar(15),
		last_modified timestamp,
		date_created timestamp,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS pages_rev (
		internal_id int NOT NULL AUTO_INCREMENT,
		wiki_page_id int,
		revision_id int,
		date_modified timestamp,
		title varchar(255) NOT NULL,
		content TEXT,
		tags TEXT,
		created_by varchar(15) NOT NULL,
		deleted int,
		last_modified_by varchar(15),
		last_modified timestamp,
		date_created timestamp,
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS queue (
		internal_id int NOT NULL AUTO_INCREMENT,
		name varchar(50),
		question TEXT,
		date_completed timestamp,
		date_created timestamp,
		assigned varchar(255),
		status varchar(60),
		PRIMARY KEY (internal_id)
		) CHARACTER SET utf8;`)
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
	if err != nil {
		log.Fatal(err)
	}
	_, err = TokenInsert.Exec(u.Username, tk.Token, tk.Expires, tk.Token, tk.Expires)
	if err != nil {
		log.Fatal(err)
	}
}

func RepairPageAndComment(w http.ResponseWriter, r *http.Request, InternalId int, c Comment) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var status int
	var newRepair = false

	err = db.QueryRow(`
	SELECT status
	FROM repairs WHERE wiki_page_id = ?`, InternalId).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		newRepair = true
	}

	if newRepair == true {
		status = 1
	} else if status == 0 {
		status = 1
	} else if status == 1 {
		status = 0
	}

	RepairPage, err := db.Prepare(`
	INSERT INTO repairs (wiki_page_id, status, username) VALUES ( ?, ?, ? ) ON DUPLICATE KEY UPDATE status = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = RepairPage.Exec(InternalId, status, c.CreatedBy, status)
	if err != nil {
		log.Fatal(err)
	}

	CommentInsert, err := db.Prepare(`
	INSERT INTO comments (title, content, wiki_page_id, created_by, date_created) VALUES ( ?, ?, ?, ?, ? )
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = CommentInsert.Exec(c.Title, c.Body, c.WikiPageId, c.CreatedBy, c.DateCreated)

	if err != nil {
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusFound)
}

func CreateComment(w http.ResponseWriter, r *http.Request, c Comment) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	CommentInsert, err := db.Prepare(`
	INSERT INTO comments (title, content, wiki_page_id, created_by, date_created) VALUES ( ?, ?, ?, ?, ? )
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = CommentInsert.Exec(c.Title, c.Body, c.WikiPageId, c.CreatedBy, c.DateCreated)
	if err != nil {
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(c.WikiPageId), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/pages/view/"+strconv.Itoa(c.WikiPageId), http.StatusFound)
}

func RepairPage(w http.ResponseWriter, r *http.Request, InternalId int, username string) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var status int
	var newRepair = false

	err = db.QueryRow(`
	SELECT status
	FROM repairs WHERE wiki_page_id = ?`, InternalId).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		newRepair = true
	}

	if newRepair == true {
		status = 1
	} else if status == 0 {
		status = 1
	} else if status == 1 {
		status = 0
	}

	RepairPage, err := db.Prepare(`
	INSERT INTO repairs (wiki_page_id, status, username) VALUES ( ?, ?, ? ) ON DUPLICATE KEY UPDATE status = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = RepairPage.Exec(InternalId, status, username, status)
	if err != nil {
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusFound)
}

func LikePage(w http.ResponseWriter, r *http.Request, InternalId int, username string) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var status int
	var newVote = false

	err = db.QueryRow(`
	SELECT status
	FROM likes WHERE wiki_page_id = ? AND username = ?`, InternalId, username).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		newVote = true
	}

	if newVote == true {
		status = 1
	} else if status == 0 {
		status = 1
	} else if status == 1 {
		status = 0
	}

	LikePage, err := db.Prepare(`
	INSERT INTO likes (wiki_page_id, username, status) VALUES ( ?, ?, ? ) ON DUPLICATE KEY UPDATE status = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = LikePage.Exec(InternalId, username, status, status)
	if err != nil {
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/pages/view/"+strconv.Itoa(InternalId), http.StatusFound)
}

func CreateCategory(w http.ResponseWriter, r *http.Request, c Tag) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	UserInsert, err := db.Prepare(`
	INSERT INTO tags (name, created_by, date_created) VALUES ( ?, ?, ? )
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = UserInsert.Exec(c.Name, c.CreatedBy, c.DateCreated)
	if err != nil {
		http.Redirect(w, r, "/categories/create/", http.StatusFound)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func FetchCategories(w http.ResponseWriter, r *http.Request) []Tag {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		tags []Tag
		name string
	)
	rows, err := db.Query("SELECT name FROM tags")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		tags = append(tags, Tag{Name: name})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return tags
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
	if err != nil {
		log.Fatal(err)
	}
	_, err = UserInsert.Exec(u.Name, u.Username, u.Email, u.Password)
	if err != nil {
		http.Redirect(w, r, "/users/create/", http.StatusFound)
	}
	http.Redirect(w, r, "/users/login/", http.StatusFound)
}

func CreateEditPreviewPage(w http.ResponseWriter, r *http.Request, s WikiPage) int {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var username, dateCreated string
	// Get original Created By
	err = db.QueryRow(`
	SELECT created_by, date_created
	FROM pages WHERE internal_id = ?`, s.InternalId).Scan(&username, &dateCreated)
	if err != nil {
		log.Fatal(err)
	}
	// Set deleted to 0 during creation.
	s.Deleted = 0
	PageInsert, err := db.Prepare(`
	INSERT INTO pages_preview (title, wiki_page_id, content, tags, created_by, deleted, last_modified, last_modified_by, date_created) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal(err)
	}
	var res sql.Result
	res, err = PageInsert.Exec(s.Title, s.InternalId, s.Content, strings.Join(s.Tags, ","), username, s.Deleted, s.LastModified, s.LastModifiedBy, dateCreated)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	wikiPageId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return int(wikiPageId)
}

func CreatePreviewPage(w http.ResponseWriter, r *http.Request, s WikiPage) int {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Set deleted to 0 during creation.
	s.Deleted = 0
	PageInsert, err := db.Prepare(`
	INSERT INTO pages_preview (title, content, tags, created_by, deleted, date_created) VALUES ( ?, ?, ?, ?, ?, ? )
	`)
	if err != nil {
		log.Fatal(err)
	}
	var res sql.Result
	res, err = PageInsert.Exec(s.Title, s.Content, strings.Join(s.Tags, ","), s.Username, s.Deleted, s.DateCreated)
	if err != nil {
		log.Fatal(err)
	}

	wikiPageId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return int(wikiPageId)
}

func CreatePage(w http.ResponseWriter, r *http.Request, InternalId int) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var title, content, tags, createdBy, existingPage, deleted, dateCreated, lastModified, lastModifiedBy, revisionId string
	// Get the original page
	err = db.QueryRow(`
	SELECT title, COALESCE(wiki_page_id, '') as wiki_page_id, content, COALESCE(tags, '') as tags, created_by, last_modified, COALESCE(last_modified_by, '') as last_modified_by, deleted, date_created
	FROM pages_preview WHERE internal_id = ?`, InternalId).Scan(&title, &existingPage, &content, &tags, &createdBy, &lastModified, &lastModifiedBy, &deleted, &dateCreated)
	if err != nil {
		log.Fatal(err)
	}
	ep, _ := strconv.Atoi(existingPage)
	d, err := strconv.Atoi(deleted)
	if err != nil {
		log.Fatal(err)
	}
	s := &WikiPage{Title: title, InternalId: ep, Content: content, Tags: strings.Split(tags, ","), LastModified: lastModified, LastModifiedBy: lastModifiedBy, Deleted: d, DateCreated: dateCreated}

	if existingPage == "" {
		PageInsert, err := db.Prepare(`
		INSERT INTO pages (title, content, tags, created_by, deleted, date_created) VALUES ( ?, ?, ?, ?, ?, ? )
		`)
		if err != nil {
			log.Fatal(err)
		}
		var res sql.Result

		res, err = PageInsert.Exec(title, content, strings.Join(s.Tags, ","), createdBy, deleted, dateCreated)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusInternalServerError)
		}

		wikiPageId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		PageRevisionInsert, err := db.Prepare(`
		INSERT INTO pages_rev (wiki_page_id, revision_id, title, content, tags, created_by, deleted, date_created, last_modified_by, last_modified)
		VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )
		`)
		if err != nil {
			log.Fatal(err)
		}
		_, err = PageRevisionInsert.Exec(wikiPageId, 1, title, content, strings.Join(s.Tags, ","), createdBy, deleted, dateCreated, lastModifiedBy, lastModified)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(int(wikiPageId)), http.StatusFound)
	} else {
		// Get the original page
		err = db.QueryRow(`
		SELECT title, content, COALESCE(tags, '') as tags, created_by, deleted, last_modified, COALESCE(last_modified_by, '') as last_modified_by, date_created
		FROM pages WHERE internal_id = ?`, existingPage).Scan(&title, &content, &tags, &createdBy, &deleted,
			&lastModified, &lastModifiedBy, &dateCreated)
		if err != nil {
			log.Fatal(err)
		}
		if title == s.Title && content == s.Content && tags == strings.Join(s.Tags, ",") {
			http.Redirect(w, r, "/pages/view/"+strconv.Itoa(ep), http.StatusFound)
			return
		}
		// Get latest revision_number
		rows, err := db.Query("SELECT revision_id FROM pages_rev WHERE wiki_page_id = ?", existingPage)
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
		if err != nil {
			log.Fatal(err)
		}
		i++
		// Insert into revisions
		PageRevisionInsert, err := db.Prepare(`
		INSERT INTO pages_rev (wiki_page_id, revision_id, title, content, tags, created_by, deleted, date_created, last_modified_by, last_modified)
		VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )
		`)
		if err != nil {
			log.Fatal(err)
		}
		_, err = PageRevisionInsert.Exec(existingPage, i, s.Title, s.Content, strings.Join(s.Tags, ","), createdBy, deleted, dateCreated, s.LastModifiedBy, s.LastModified)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusInternalServerError)
		}

		// Set deleted to 0 for newly updated.
		s.Deleted = 0

		PageUpdate, err := db.Prepare(`
		UPDATE pages SET title = ?, content = ?, tags = ?, deleted = ?, last_modified = ?, last_modified_by = ?
		WHERE internal_id = ?
		`)
		if err != nil {
			log.Fatal(err)
		}
		_, err = PageUpdate.Exec(s.Title, s.Content, strings.Join(s.Tags, ","), s.Deleted, s.LastModified, s.LastModifiedBy, s.InternalId)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		http.Redirect(w, r, "/pages/view/"+strconv.Itoa(ep), http.StatusFound)

	}

}

func RestorePage(w http.ResponseWriter, r *http.Request, InternalId int) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	PageUpdate, err := db.Prepare(`
	UPDATE pages SET deleted = ?
	WHERE internal_id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = PageUpdate.Exec(0, InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func RollbackPage(w http.ResponseWriter, r *http.Request, RollbackId int) string {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var title, content, tags, username, lastModifiedBy, deleted, lastModified, dateCreated, wikiPageId string

	// Get the revision page
	err = db.QueryRow(`
	SELECT title, content, COALESCE(tags, '') as tags, created_by, deleted, last_modified, COALESCE(last_modified_by, '') as last_modified_by, date_created, wiki_page_id
	FROM pages_rev WHERE internal_id = ?`, RollbackId).Scan(&title, &content, &tags, &username, &deleted,
		&lastModified, &lastModifiedBy, &dateCreated, &wikiPageId)
	if err != nil {
		log.Fatal(err)
	}

	// Update original
	PageUpdate, err := db.Prepare(`
	UPDATE pages SET title = ?, content = ?, tags = ?, created_by = ?, deleted = ?, last_modified = ?, last_modified_by = ?, date_created = ?
	WHERE internal_id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = PageUpdate.Exec(title, content, tags, username, deleted, lastModified, lastModifiedBy, dateCreated, wikiPageId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	return wikiPageId
}

func SearchWikiPages(w http.ResponseWriter, r *http.Request, searchKey string) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var wikiPages []WikiPage
	var title, content, deleted, username, lastModifiedBy, tags, internalId, lastModified, dateCreated string

	rows, err := db.Query(`
	SELECT internal_id, title, deleted, content, COALESCE(tags, '') as tags, created_by, COALESCE(last_modified_by, '') as last_modified_by, last_modified, date_created
	FROM pages WHERE content REGEXP ? OR title REGEXP ?
	`, searchKey, searchKey)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&internalId, &title, &deleted, &content, &tags, &username, &lastModifiedBy,
			&lastModified, &dateCreated)
		if err != nil {
			log.Fatal(err)
		}
		deletedId, err := strconv.Atoi(deleted)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusInternalServerError)
		}
		id, err := strconv.Atoi(internalId)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusInternalServerError)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Title: title, Deleted: deletedId, Content: content, Tags: strings.Split(tags, ","), DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, CreatedBy: username})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	return wikiPages
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
	if err != nil {
		log.Fatal(err)
	}
	_, err = PageUpdate.Exec(1, InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func ShowRevisionPage(w http.ResponseWriter, r *http.Request, InternalId int) (*WikiPageRevision, *WikiPage) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	defer db.Close()
	var title, content, tags, dateCreated, lastModified, lastModifiedBy, username, revisionId, wikiPageId string
	err = db.QueryRow(`
	SELECT title, content, COALESCE(tags, '') as tags, wiki_page_id, revision_id, date_created, last_modified, COALESCE(last_modified_by, '') as last_modified_by, created_by FROM pages_rev WHERE internal_id = ?
	`, InternalId).Scan(&title, &content, &tags, &wikiPageId, &revisionId, &dateCreated, &lastModified, &lastModifiedBy, &username)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	id, err := strconv.Atoi(revisionId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	wid, err := strconv.Atoi(wikiPageId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	wpr := &WikiPageRevision{Title: title, WikiPageId: wid, Content: content, Tags: strings.Split(tags, ","), RevisionId: id, DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, CreatedBy: username}

	id, err = strconv.Atoi(wikiPageId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	err = db.QueryRow(`
	SELECT title, content, COALESCE(tags, '') as tags, date_created, last_modified, COALESCE(last_modified_by, '') as last_modified_by, created_by FROM pages WHERE internal_id = ?
	`, id).Scan(&title, &content, &tags, &dateCreated, &lastModified, &lastModifiedBy, &username)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	wp := &WikiPage{Title: title, Content: content, Tags: strings.Split(tags, ","), DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, CreatedBy: username}

	return wpr, wp
}

func ShowPage(w http.ResponseWriter, r *http.Request, InternalId int) (*WikiPage, error) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var title, content, tags, deleted, dateCreated, lastModified, lastModifiedBy, username string

	err = db.QueryRow(`
	SELECT title, content, COALESCE(tags, '') as tags, deleted, date_created, last_modified, COALESCE(last_modified_by, '') as last_modified_by, created_by FROM pages WHERE internal_id = ?
	`, InternalId).Scan(&title, &content, &tags, &deleted, &dateCreated, &lastModified, &lastModifiedBy, &username)
	if len(tags) == 0 {
		tags = "None"
	}
	deletedString, _ := strconv.Atoi(deleted)

	return &WikiPage{Title: title, Content: content, Deleted: deletedString, Tags: strings.Split(tags, ","), DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, CreatedBy: username}, err
}

func LoadAllPagesToday(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		id             int
		title          string
		createdBy      string
		dateCreated    string
		lastModifiedBy string
		lastModified   string
	)

	rows, err := db.Query("SELECT internal_id, title, created_by, date_created, COALESCE(last_modified_by, '') as last_modified_by, last_modified FROM pages WHERE deleted = ? AND date_created > DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 1 DAY) ORDER BY date_created DESC", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title, &createdBy, &dateCreated, &lastModifiedBy, &lastModified)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Title: title, DateCreated: dateCreated, CreatedBy: createdBy, LastModifiedBy: lastModifiedBy, LastModified: lastModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages

}

func LoadPageLast25(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		id             int
		title          string
		createdBy      string
		dateCreated    string
		lastModifiedBy string
		lastModified   string
	)
	rows, err := db.Query("SELECT internal_id, title, created_by, date_created, COALESCE(last_modified_by, '') as last_modified_by, last_modified FROM pages WHERE deleted = ? ORDER BY last_modified DESC, date_created DESC LIMIT 25 ", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title, &createdBy, &dateCreated, &lastModifiedBy, &lastModified)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Title: title, DateCreated: dateCreated, CreatedBy: createdBy, LastModifiedBy: lastModifiedBy, LastModified: lastModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func LoadTop5Voted(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages []WikiPage
		id        int
		deleted   int
		title     string
		votes     int
	)
	rows, err := db.Query("select pages.internal_id,pages.deleted,pages.title,count(*) as count from pages join likes on pages.internal_id = likes.wiki_page_id where deleted = ? and status = ? group by pages.internal_id order by count desc limit 5;", 0, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &deleted, &title, &votes)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Deleted: deleted, Title: title, Liked: votes})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func LoadMyVoted(w http.ResponseWriter, r *http.Request, username string) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		id             int
		deleted		   int
		title          string
		createdBy      string
		tags		   string
		dateCreated    string
		lastModifiedBy string
		lastModified   string
	)
	rows, err := db.Query("select pages.internal_id,pages.deleted,pages.created_by,COALESCE(tags, '') as tags,pages.date_created,COALESCE(last_modified_by, '') as last_modified_by,pages.last_modified,pages.title from pages join likes on pages.internal_id = likes.wiki_page_id where pages.deleted = ? and likes.status = ? and likes.username = ? group by pages.internal_id ORDER BY pages.last_modified DESC, pages.date_created DESC", 0, 1, username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &deleted, &createdBy, &tags, &dateCreated, &lastModifiedBy, &lastModified, &title)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Deleted: deleted, Title: title, Tags: strings.Split(tags, ","), DateCreated: dateCreated, CreatedBy: createdBy, LastModifiedBy: lastModifiedBy, LastModified: lastModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func LoadNeedsImprovement(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		id             int
		deleted		   int
		title          string
		createdBy      string
		tags		   string
		dateCreated    string
		lastModifiedBy string
		lastModified   string
	)
	rows, err := db.Query("select pages.internal_id,repairs.username,pages.deleted,COALESCE(tags, '') as tags,pages.date_created,COALESCE(last_modified_by, '') as last_modified_by,pages.last_modified,pages.title from pages join repairs on pages.internal_id = repairs.wiki_page_id where pages.deleted = ? and repairs.status = ? group by pages.internal_id ORDER BY pages.last_modified DESC, pages.date_created DESC", 0, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &createdBy, &deleted, &tags, &dateCreated, &lastModifiedBy, &lastModified, &title)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Deleted: deleted, Title: title, Tags: strings.Split(tags, ","), DateCreated: dateCreated, CreatedBy: createdBy, LastModifiedBy: lastModifiedBy, LastModified: lastModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func DownloadAllPages(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		title          string
		content		   string
	)
	rows, err := db.Query("SELECT title, content FROM pages WHERE deleted = ? ORDER BY internal_id DESC", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&title, &content)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{Title: title, Content: content})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func LoadAllPages(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages      []WikiPage
		id             int
		title          string
		createdBy      string
		tags		   string
		dateCreated    string
		lastModifiedBy string
		lastModified   string
	)
	rows, err := db.Query("SELECT internal_id, title, created_by, COALESCE(tags, '') as tags, date_created, COALESCE(last_modified_by, '') as last_modified_by, last_modified FROM pages WHERE deleted = ? ORDER BY last_modified DESC, date_created DESC", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title, &createdBy, &tags, &dateCreated, &lastModifiedBy, &lastModified)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, Title: title, Tags: strings.Split(tags, ","), DateCreated: dateCreated, CreatedBy: createdBy, LastModifiedBy: lastModifiedBy, LastModified: lastModified})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func Top10Commented(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages    []WikiPage
		id           int
		commentCount int
		title        string
		createdBy    string
		dateCreated  string
	)
	rows, err := db.Query("SELECT comments.wiki_page_id, count(comments.wiki_page_id) as comment_count, pages.title, comments.created_by, max(comments.date_created) FROM comments, pages WHERE pages.internal_id = wiki_page_id AND deleted = ? GROUP BY comments.wiki_page_id ORDER BY max(comments.date_created) DESC limit 10", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &commentCount, &title, &createdBy, &dateCreated)
		if err != nil {
			log.Fatal(err)
		}
		wikiPages = append(wikiPages, WikiPage{InternalId: id, CommentCount: commentCount, Title: title, DateCreated: dateCreated, CreatedBy: createdBy})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return wikiPages
}

func ShowPreviewPage(w http.ResponseWriter, r *http.Request, InternalId int) (*WikiPage, error) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var title, content, tags, dateCreated, lastModified, lastModifiedBy, username string

	err = db.QueryRow(`
	SELECT title, content, tags, date_created, last_modified, COALESCE(last_modified_by, '') as last_modified_by, created_by FROM pages_preview WHERE internal_id = ?
	`, InternalId).Scan(&title, &content, &tags, &dateCreated, &lastModified, &lastModifiedBy, &username)

	tagsArray := strings.Split(tags, ",")
	return &WikiPage{Title: title, Content: content, Tags: tagsArray, DateCreated: dateCreated, LastModified: lastModified, LastModifiedBy: lastModifiedBy, CreatedBy: username}, err
}

func FetchDeletedPages(w http.ResponseWriter, r *http.Request) []WikiPage {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		wikiPages []WikiPage
		id        int
		title     string
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
	if err != nil {
		log.Fatal(err)
	}
	var (
		wikiPages      []WikiPageRevision
		revisionId     int
		title          string
		dateModified   string
		lastModifiedBy string
	)
	rows, err := db.Query(`SELECT internal_id, revision_id, title, date_modified, COALESCE(last_modified_by, '') as last_modified_by
		FROM pages_rev WHERE wiki_page_id = ?`, internalId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&internalId, &revisionId, &title, &dateModified, &lastModifiedBy)
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

func GetRepairsForPage(w http.ResponseWriter, r *http.Request, internalId int) int {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var status int

	err = db.QueryRow(`
	SELECT status FROM repairs where wiki_page_id = ?`, internalId).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		return 2
	}
	return status
}

func GetLikeForPagePerUser(w http.ResponseWriter, r *http.Request, internalId int, username string) int {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var status int

	err = db.QueryRow(`
	SELECT status FROM likes where wiki_page_id = ? and username = ?
	`, internalId, username).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		return 2
	}
	return status
}

func FetchComments(w http.ResponseWriter, r *http.Request, internalId int) []Comment {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		comment     []Comment
		createdBy   string
		dateCreated string
		title       string
		content     string
	)

	rows, err := db.Query(`SELECT title, content, created_by, date_created FROM comments where wiki_page_id = ? ORDER BY date_created DESC`, internalId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&title, &content, &createdBy, &dateCreated)
		if err != nil {
			log.Fatal(err)
		}
		comment = append(comment, Comment{Title: title, Body: content, CreatedBy: createdBy, DateCreated: dateCreated})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return comment
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
