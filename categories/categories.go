package categories

import (
	"bytes"
	"github.com/zerodayz/gowiki/database"
	"github.com/zerodayz/gowiki/pages"
	"html/template"
	"net/http"
	"time"
)

var (
	templatePath = "tmpl/categories/"
)

// Handlers

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	s := database.WikiPage{}
	c := database.Tag{}

	t := template.Must(template.ParseFiles(templatePath + "create.html"))
	username := pages.ReadCookie(w, r)
	s.UserLoggedIn = username

	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	// Set username to Logged in User.
	s.Username = username
	if r.Method == "POST" {
		r.ParseForm()
		c.Name = r.PostFormValue("name")
		c.CreatedBy = s.Username
		date := time.Now().UTC()
		c.DateCreated = date.Format("20060102150405")
		database.CreateCategory(w, r, c)
	}

	bufCategories := bytes.NewBuffer(nil)
	existingCategories := database.FetchCategories(w, r)
	bufCategories.Write([]byte(`<div class="header-text"><h1>Create Keepsake Category</h1></div>`))
	if len(existingCategories) == 0 {
		bufCategories.Write([]byte(`There are no categories yet.`))
	} else {
		bufCategories.Write([]byte(`Existing Categories: `))
		for _, f := range existingCategories {
			bufCategories.Write([]byte(`<div class="categories">` + f.Name + `</div>`))
		}
	}
	s.DisplayComment = template.HTML(bufCategories.String())

	err := t.ExecuteTemplate(w, "create.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
