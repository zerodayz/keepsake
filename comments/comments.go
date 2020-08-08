package comments

import (
	"github.com/zerodayz/gowiki/database"
	"github.com/zerodayz/gowiki/pages"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	c := database.Comment{}
	username := pages.ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	// Set username to Logged in User.
	c.CreatedBy = username

	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		c.WikiPageId = id

		c.Title = r.PostFormValue("comment_title")
		c.Body = r.PostFormValue("comment_message")
		date := time.Now().UTC()
		c.DateCreated = date.Format("20060102150405")

		database.CreateComment(w, r, c)
	}
}

var validPath = regexp.MustCompile("^/(comments)/(create)/([0-9]+)$")

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[3])
	}
}
