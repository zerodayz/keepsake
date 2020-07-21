package pages

import (
	"bytes"
	"github.com/zerodayz/gowiki/database"
	"gitlab.com/golang-commonmark/markdown"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Page struct {
	Title          string
	InternalId     string
	UserLoggedIn   string
	CreatedBy      string
	DateCreated    string
	LastModified   string
	LastModifiedBy string
	EditTitle      string
	Body           string
	DisplayBody    template.HTML
	Errors         map[string]string
}

var (
	templatePath = "tmpl/pages/"
	datapath     = "data/pages/"
	templates    = template.Must(template.ParseFiles(
		templatePath+"edit.html", templatePath+"view.html", templatePath+"search.html"))
)

func max(x int, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}
func min(x int, y int) int {
	if x > y {
		return y
	} else {
		return x
	}
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ReadCookie(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("gowiki_session")
	if err != nil {
		return "Unauthorized"
	} else {
		value := c.Value
		username := database.GetSessionKey(w, r, value)
		return username
	}
}

func (p *Page) Validate(w http.ResponseWriter, r *http.Request) bool {
	p.Errors = make(map[string]string)
	if len(p.EditTitle) != 0 {
		var validFilename = regexp.MustCompile("^([a-z0-9_]+)$")

		match := validFilename.Match([]byte(p.EditTitle))
		if match == false {
			p.Errors["Title"] = "Please enter a valid title. Allowed charset: [a-z0-9_]"
		}
	}
	if p.EditTitle != p.Title {
		exists := Exists(datapath + p.EditTitle + ".md")
		if exists == true {
			p.Errors["Title"] = "Unable to save. Another Wiki page already exists with the requested name."
		}
	}
	if len(p.Body) == 0 {
		p.Errors["Content"] = "Unable to save with empty body."
	}
	if len(p.EditTitle) == 0 && len(p.Body) == 0 {
		http.Redirect(w, r, "/pages/view/"+p.Title, http.StatusFound)
	}
	return len(p.Errors) == 0
}

func LoadPage(w http.ResponseWriter, r *http.Request, InternalId int) (*Page, error) {
	s := database.ShowPage(w, r, InternalId)
	return &Page{Title: s.Title, Body: s.Content, InternalId: strconv.Itoa(InternalId), CreatedBy: s.Username, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated}, nil
}

func LoadRevisionPage(w http.ResponseWriter, r *http.Request, InternalId int) (*Page, error) {
	s := database.ShowRevisionPage(w, r, InternalId)
	return &Page{Title: s.Title, Body: s.Content, InternalId: strconv.Itoa(InternalId), CreatedBy: s.Username, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated}, nil
}

// Handlers

func ViewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	username := ReadCookie(w, r)
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	p, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/pages/create", http.StatusNotFound)
		return
	}
	md := markdown.New()
	p.DisplayBody = template.HTML(md.RenderToString([]byte(p.Body)))

	p.UserLoggedIn = username
	RenderTemplate(w, "view", p)
}

func RevisionsViewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	t := template.Must(template.ParseFiles(templatePath + "revision.html"))
	username := ReadCookie(w, r)
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	p, err := LoadRevisionPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	md := markdown.New()
	p.DisplayBody = template.HTML(md.RenderToString([]byte(p.Body)))

	p.UserLoggedIn = username
	err = t.ExecuteTemplate(w, "revision.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func EditHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}
	p, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	p.UserLoggedIn = username

	RenderTemplate(w, "edit", p)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	s := database.WikiPage{}
	p := &Page{}
	t := template.Must(template.ParseFiles(templatePath + "create.html"))

	username := ReadCookie(w, r)
	p.UserLoggedIn = username

	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	// Set username to Logged in User.
	s.Username = username
	if r.Method == "POST" {
		r.ParseForm()
		s.Title = r.PostFormValue("title")
		s.Content = r.PostFormValue("body")
		date := time.Now().UTC()
		s.DateCreated = date.Format("20060102150405")

		database.CreatePage(w, r, s)
	}

	err := t.ExecuteTemplate(w, "create.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	database.DeletePage(w, r, id)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	s := database.WikiPage{}
	p := &Page{}
	username := ReadCookie(w, r)

	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	p.UserLoggedIn = username
	// Set username to Logged in User.
	s.Username = username
	if r.Method == "POST" {
		r.ParseForm()
		s.InternalId = id
		s.Title = r.PostFormValue("title")
		s.Content = r.PostFormValue("body")
		date := time.Now().UTC()
		s.LastModified = date.Format("20060102150405")
		s.LastModifiedBy = username

		database.UpdatePage(w, r, s)
	}

	http.Redirect(w, r, "/pages/view/"+InternalId, http.StatusFound)
}

//func SearchHandler(w http.ResponseWriter, r *http.Request) {
//	username := ReadCookie(w, r)
//
//	u, err := url.Parse(r.URL.String())
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	params := u.Query()
//	searchKey := params.Get("q")
//	var fileReg = regexp.MustCompile(`^[a-z0-9_]+\.md$`)
//	var searchQuery = regexp.MustCompile(searchKey)
//
//	buf := bytes.NewBuffer(nil)
//
//	files, err := ioutil.ReadDir(datapath)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	if len(searchKey) == 0 {
//		return
//	}
//	buf.Write([]byte(`<div id="items"></div>`))
//	for _, f := range files {
//		if fileReg.MatchString(f.Name()) {
//			content, err := ioutil.ReadFile(datapath + f.Name())
//			var contentLenght = len(content)
//			if err != nil {
//				http.Error(w, err.Error(), http.StatusInternalServerError)
//				return
//			}
//
//			var indexes = searchQuery.FindAllIndex(content, -1)
//			if len(indexes) != 0 {
//				var occurences = strconv.Itoa(len(indexes))
//				fileName := strings.Split(f.Name(), ".")
//				if username == "Unauthorized" {
//					buf.Write([]byte(`
//					<div class="found">Found ` + occurences + ` occurrences.
//					<a href="/pages/view/` + fileName[0] + `"><img src="/lib/icons/public-24px.svg"></a>
//					<label for="search-content" class="search-collapsible">
//					` + fileName[0] + `</label>
//					<div id="search-content" class="search-content">`))
//				} else {
//					buf.Write([]byte(`
//					<div class="found">Found ` + occurences + ` occurrences.
//					<a href="/pages/view/` + fileName[0] + `"><img src="/lib/icons/public-24px.svg"></a>
//					<a href="/pages/edit/` + fileName[0] + `"><img src="/lib/icons/edit-black-24px.svg"></a>
//					<label for="search-content" class="search-collapsible">
//					` + fileName[0] + `</label>
//					<div id="search-content" class="search-content">`))
//				}
//				for _, k := range indexes {
//					var start = k[0]
//					var end = k[1]
//
//					var showStart = max(start-100, 0)
//					var showEnd = min(end+100, contentLenght-1)
//
//					for !utf8.RuneStart(content[showStart]) {
//						showStart = max(showStart-1, 0)
//					}
//					for !utf8.RuneStart(content[showEnd]) {
//						showEnd = min(showEnd-1, contentLenght)
//					}
//					buf.Write([]byte(`<pre><code>`))
//					buf.WriteString(template.HTMLEscapeString(string(content[showStart:start])))
//					buf.Write([]byte(`<b>`))
//					buf.WriteString(template.HTMLEscapeString(string(content[start:end])))
//					buf.Write([]byte(`</b>`))
//					if (end - 1) != showEnd {
//						buf.WriteString(template.HTMLEscapeString(string(content[end:showEnd])))
//					}
//					buf.Write([]byte(`</code></pre>`))
//				}
//				buf.Write([]byte(`</div></div>`))
//				buf.WriteByte('\n')
//			}
//		}
//	}
//
//	p := &Page{Title: searchKey, Body: []byte(buf.String()), UserLoggedIn: username}
//	p.DisplayBody = template.HTML(buf.String())
//	p.Title = searchKey
//	p.UserLoggedIn = username
//	RenderTemplate(w, "search", p)
//}

func RevisionsHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	p := Page{}
	t := template.Must(template.ParseFiles(templatePath + "revisions.html"))

	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, _ := strconv.Atoi(InternalId)

	buf := bytes.NewBuffer(nil)
	wikiRevisionPages, wikiPage := database.FetchRevisionPages(w, r, id)
	p.Title = wikiPage

	buf.Write([]byte(`<div>There are ` + strconv.Itoa(len(wikiRevisionPages)) + ` revision(s) available.</div>`))
	for _, f := range wikiRevisionPages {
		buf.Write([]byte(`<b>` + f.Title  + `</b><br><a href="/revisions/view/` + strconv.Itoa(f.InternalId) + `">Revision ` + strconv.Itoa(f.RevisionId) + `</a> | ` + `Modified by ` +
			f.LastModifiedBy + ` | ` + f.LastModified ))
		buf.Write([]byte(`<br>`))
	}

	p.DisplayBody = template.HTML(buf.String())
	p.UserLoggedIn = username

	err := t.ExecuteTemplate(w, "revisions.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func RecycleBinHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}
	t := template.Must(template.ParseFiles(templatePath + "trash.html"))

	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	buf := bytes.NewBuffer(nil)
	wikiPages := database.FetchDeletedPages(w, r)

	buf.Write([]byte(`<div>There are ` + strconv.Itoa(len(wikiPages)) + ` files in Recycle Bin.</div>`))
	for _, f := range wikiPages {
		buf.Write([]byte(`<b>` + f.Title + ` </b><br>
			<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` + ` | ` +
			`<a href="/pages/restore/` + strconv.Itoa(f.InternalId) + `">Remove from Bin</a>`))
		buf.Write([]byte(`<br>`))
	}

	p.DisplayBody = template.HTML(buf.String())
	p.UserLoggedIn = username
	err := t.ExecuteTemplate(w, "trash.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RestoreHandler(w http.ResponseWriter, r *http.Request) {
	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	var restorePath = regexp.MustCompile("^/pages/restore/([a-z0-9_-]+)$")
	m := restorePath.FindStringSubmatch(r.URL.Path)
	fileName := m[1]
	newFileName := strings.Split(fileName, "-")

	err := os.Rename(datapath+"deleted/"+fileName+".md", datapath+newFileName[0]+".md")
	if err != nil {
		http.Redirect(w, r, "/pages/trash", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/pages/view/"+newFileName[0], http.StatusFound)
	return
}

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(pages|revisions)/(edit|save|view|delete|revisions)/([0-9]+)$")

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
