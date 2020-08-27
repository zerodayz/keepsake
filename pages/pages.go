package pages

import (
	"bytes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/zerodayz/gowiki/database"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

var (
	templatePath = "tmpl/pages/"
	datapath     = "data/pages/"
	templates    = template.Must(template.ParseFiles(
		templatePath+"edit.html", templatePath+"view.html", templatePath+"preview.html", templatePath+"search.html"))
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

func LoadPage(w http.ResponseWriter, r *http.Request, InternalId int) (*database.WikiPage, error) {
	s := database.ShowPage(w, r, InternalId)
	return &database.WikiPage{Title: s.Title, Body: s.Content, InternalId: InternalId, CreatedBy: s.CreatedBy, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated}, nil
}

func LoadPreviewPage(w http.ResponseWriter, r *http.Request, InternalId int) (*database.WikiPage, error) {
	s := database.ShowPreviewPage(w, r, InternalId)
	return &database.WikiPage{Title: s.Title, Body: s.Content, InternalId: InternalId, CreatedBy: s.CreatedBy, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated}, nil
}

func LoadRevisionPage(w http.ResponseWriter, r *http.Request, InternalId int) (*database.WikiPageRevision, *database.WikiPage) {
	wpr, wp := database.ShowRevisionPage(w, r, InternalId)
	return &database.WikiPageRevision{Title: wpr.Title, WikiPageId: wpr.WikiPageId, RevisionId: wpr.RevisionId, Content: wpr.Content, InternalId: InternalId, CreatedBy: wpr.CreatedBy, LastModified: wpr.LastModified, LastModifiedBy: wpr.LastModifiedBy, DateCreated: wpr.DateCreated},
	wp
}

// Handlers

func ViewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	t := template.Must(template.ParseFiles(templatePath + "view.html"))

	username := ReadCookie(w, r)
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	s, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/pages/create", http.StatusNotFound)
		return
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	var buf bytes.Buffer
	md.Convert([]byte(s.Body), &buf)
	s.DisplayBody = template.HTML(buf.String())

	s.UserLoggedIn = username

	// Fetch comments
	var bufComments bytes.Buffer

	comments := database.FetchComments(w, r, id)
	bufComments.Write([]byte(`<div>There are ` + strconv.Itoa(len(comments)) + ` comment(s).</div>`))
	for _, f := range comments {
		bufComments.Write([]byte(`
				<div class="found">Comment by ` + f.CreatedBy + ` on ` + f.DateCreated + `
				<label for="search-content" class="search-collapsible">`+ f.Title +`</label>
				<div id="search-content" class="comment-content">
				<pre><code>` + f.Body + `</code></pre></div></div>`))
	}
	s.DisplayComment = template.HTML(bufComments.String())

	err = t.ExecuteTemplate(w, "view.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "dashboard.html"))

	username := ReadCookie(w, r)
	dateYesterday := time.Now().AddDate(0, 0, -1).UTC()

	bufComment := bytes.NewBuffer(nil)
	wikiPagesTop10Commented := database.Top10Commented(w, r)
	bufComment.Write([]byte(`<div class="header-text"><h1>Keepsake Last 10 Discussed</h1></div>`))
	if len(wikiPagesTop10Commented) == 0 {
		bufComment.Write([]byte(`There are no discussions.`))
	} else {
		for _, f := range wikiPagesTop10Commented {
			// 2020-08-02 23:44:28
			dateCreated := time.Now()
			comments := database.FetchComments(w, r, f.InternalId)

			dateCreated, err := time.Parse("2006-01-02 15:04:05", f.DateCreated)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if dateCreated.After(dateYesterday) {
				bufComment.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <img src="/lib/icons/comment-24px.svg" alt="New comment!"/> | Comments: ` + strconv.Itoa(len(comments)) + ` | Last commented on ` + f.DateCreated + ` by ` + f.CreatedBy + `</div>`))
			} else {
				bufComment.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> | Comments: ` + strconv.Itoa(len(comments)) + ` | Commented on ` + f.DateCreated + ` by ` + f.CreatedBy + `</div>`))
			}
		}
	}


	buf := bytes.NewBuffer(nil)
	wikiPages := database.LoadPageLast25(w, r)
	buf.Write([]byte(`<div class="header-text-n"><h1>Keepsake Last 25 Updated</h1></div>`))
	if len(wikiPages) == 0 {
		buf.Write([]byte(`There are no wiki pages.`))
	} else {
		for _, f := range wikiPages {
			// 2020-08-02 23:44:28
			dateCreated := time.Now()
			comments := database.FetchComments(w, r, f.InternalId)

			dateCreated, err := time.Parse("2006-01-02 15:04:05", f.DateCreated)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if f.LastModifiedBy == "" {
				if dateCreated.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <img src="/lib/icons/fiber_new-24px.svg" alt="New!"/> | Comments: ` + strconv.Itoa(len(comments)) + ` | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						` | Not yet modified.</div>`))
				} else {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> | Comments: ` + strconv.Itoa(len(comments)) + ` | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						` | Not yet modified.</div>`))
				}
			} else {
				dateModified, err := time.Parse("2006-01-02 15:04:05", f.LastModified)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				if dateCreated.After(dateYesterday) && dateModified.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <img src="/lib/icons/fiber_new-24px.svg" alt="New!"/> <img src="/lib/icons/new_releases-24px.svg" alt="Updated!"/> | Comments: ` + strconv.Itoa(len(comments)) + ` | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						` | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				} else if dateModified.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <img src="/lib/icons/new_releases-24px.svg" alt="Updated!"/> | Comments: ` + strconv.Itoa(len(comments)) + ` | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						` | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				} else {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> | Comments: ` + strconv.Itoa(len(comments)) + ` | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						` | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				}
			}
		}
	}

	p.DisplayBody = template.HTML(buf.String())
	p.DisplayComment = template.HTML(bufComment.String())

	p.UserLoggedIn = username
	err := t.ExecuteTemplate(w, "dashboard.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RevisionsViewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	t := template.Must(template.ParseFiles(templatePath + "revision.html"))

	username := ReadCookie(w, r)
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	wpr, wp := LoadRevisionPage(w, r, id)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	var bufferWikiPageRevision bytes.Buffer
	md.Convert([]byte(wpr.Content), &bufferWikiPageRevision)
	wpr.DisplayBody = template.HTML(bufferWikiPageRevision.String())

	var bufferWikiPage bytes.Buffer
	md.Convert([]byte(wp.Content), &bufferWikiPage)
	wp.DisplayBody = template.HTML(bufferWikiPage.String())

	wpr.UserLoggedIn = username
	err = t.ExecuteTemplate(w, "revision.html",
		struct{WikiPageRevision, WikiPage interface{}}{wpr, wp})
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
	s, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	s.UserLoggedIn = username

	RenderTemplate(w, "edit", s)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	s := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "create.html"))
	username := ReadCookie(w, r)
	s.UserLoggedIn = username

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

		s.InternalId = database.CreatePreviewPage(w, r, s)
		http.Redirect(w, r, "/preview/view/"+strconv.Itoa(s.InternalId), http.StatusFound)

	}

	err := t.ExecuteTemplate(w, "create.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PreviewCreateHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
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

	database.CreatePage(w, r, id)
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

func PreviewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	t := template.Must(template.ParseFiles(templatePath + "preview.html"))
	username := ReadCookie(w, r)
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	s, err := LoadPreviewPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/pages/create", http.StatusNotFound)
		return
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	var buf bytes.Buffer

	md.Convert([]byte(s.Body), &buf)
	s.DisplayBody = template.HTML(buf.String())

	s.UserLoggedIn = username
	err = t.ExecuteTemplate(w, "preview.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SaveHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	s := database.WikiPage{}
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

	s.UserLoggedIn = username
	if r.Method == "POST" {
		nlcr := regexp.MustCompile("\r\n")
		r.ParseForm()
		s.InternalId = id
		s.Title = r.PostFormValue("title")
		s.Content = r.PostFormValue("body")
		s.Content = string(nlcr.ReplaceAllFunc([]byte(s.Content), func(s []byte) []byte {
			return []byte("\n")
		}))
		date := time.Now().UTC()
		s.LastModified = date.Format("20060102150405")
		s.LastModifiedBy = username


		previewId := database.CreateEditPreviewPage(w, r, s)
		http.Redirect(w, r, "/preview/view/"+strconv.Itoa(previewId), http.StatusFound)

	}
}

func RevisionRollbackHandler(w http.ResponseWriter, r *http.Request, RollbackId string) {
	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(RollbackId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	internalId := database.RollbackPage(w, r, id)
	http.Redirect(w, r, "/pages/view/"+internalId, http.StatusFound)

}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	username := ReadCookie(w, r)
	t := template.Must(template.ParseFiles(templatePath + "search.html"))

	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := u.Query()
	searchKey := params.Get("q")
	var searchQuery = regexp.MustCompile(`(?i)` + searchKey)

	buf := bytes.NewBuffer(nil)


	if len(searchKey) == 0 {
		buf.Write([]byte(`
					<div class="found">Please search for something else than empty.</div>`))
	} else
	{
		buf.Write([]byte(`<div id="items"></div>`))

		s := database.SearchWikiPages(w, r, searchKey)
		for _, f := range s {
			var contentLength = len(f.Content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var indexes = searchQuery.FindAllIndex([]byte(f.Content), -1)
			if len(indexes) != 0 {
				var occurrences = strconv.Itoa(len(indexes))
				if username == "Unauthorized" {
					buf.Write([]byte(`
					<div class="found">Found ` + occurrences + ` occurrences.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` +
						`<label for="search-content" class="search-collapsible">
					` + f.Title + `</label>
					<div id="search-content" class="search-content">`))
				} else {
					buf.Write([]byte(`
					<div class="found">Found ` + occurrences + ` occurrences.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` + ` | ` +
						`<a href="/pages/edit/` + strconv.Itoa(f.InternalId) + `">Edit Page</a>
					<label for="search-content" class="search-collapsible">
					` + f.Title + `</label>
					<div id="search-content" class="search-content">`))
				}
				for _, k := range indexes {
					var start = k[0]
					var end = k[1]

					var showStart = max(start-100, 0)
					var showEnd = min(end+100, contentLength-1)

					for !utf8.RuneStart(f.Content[showStart]) {
						showStart = max(showStart-1, 0)
					}
					for !utf8.RuneStart(f.Content[showEnd]) {
						showEnd = min(showEnd-1, contentLength)
					}
					buf.Write([]byte(`<pre><code>`))
					buf.WriteString(template.HTMLEscapeString(f.Content[showStart:start]))
					buf.Write([]byte(`<b>`))
					buf.WriteString(template.HTMLEscapeString(f.Content[start:end]))
					buf.Write([]byte(`</b>`))
					if (end - 1) != showEnd {
						buf.WriteString(template.HTMLEscapeString(f.Content[end:showEnd]))
					}
					buf.Write([]byte(`</code></pre>`))
				}
				buf.Write([]byte(`</div></div>`))
				buf.WriteByte('\n')
			} else {
				if username == "Unauthorized" {
					buf.Write([]byte(`
					<div class="found">Matched title of the wiki page.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` +
						`<label for="search-content" class="search-no-collapsible">
					` + f.Title + `</label>`))
				} else {
					buf.Write([]byte(`
					<div class="found">Matched title of the wiki page.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` + ` | ` +
						`<a href="/pages/edit/` + strconv.Itoa(f.InternalId) + `">Edit Page</a>
					<label for="search-content" class="search-no-collapsible">
					` + f.Title + `</label>`))
				}
				buf.Write([]byte(`</div>`))
				buf.WriteByte('\n')
			}
		}
	}
	p := database.WikiPage{}
	p.DisplayBody = template.HTML(buf.String())
	p.Title = searchKey
	p.UserLoggedIn = username

	err = t.ExecuteTemplate(w, "search.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RevisionsHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "revisions.html"))

	username := ReadCookie(w, r)
	id, _ := strconv.Atoi(InternalId)

	buf := bytes.NewBuffer(nil)
	wikiRevisionPages, wikiPage := database.FetchRevisionPages(w, r, id)
	p.Title = wikiPage

	buf.Write([]byte(`<div>There are ` + strconv.Itoa(len(wikiRevisionPages)) + ` revision(s) available.</div>`))
	for _, f := range wikiRevisionPages {
		if len(wikiRevisionPages) == f.RevisionId {
			buf.Write([]byte(`<b>` + f.Title + `</b><br><a href="/revisions/view/` + strconv.Itoa(f.InternalId) + `">Current version </a> | ` + `Last Modified by ` +
				f.LastModifiedBy + ` on ` + f.LastModified ))
			buf.Write([]byte(`<br>`))
		} else {
			buf.Write([]byte(`<b>` + f.Title + `</b><br><a href="/revisions/view/` + strconv.Itoa(f.InternalId) + `">Revision ` + strconv.Itoa(f.RevisionId) + `</a> | ` + `Last Modified by ` +
				f.LastModifiedBy + ` on ` + f.LastModified ))
			buf.Write([]byte(`<br>`))
		}
	}

	p.DisplayBody = template.HTML(buf.String())
	p.UserLoggedIn = username

	err := t.ExecuteTemplate(w, "revisions.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func RecycleBinHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
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

func RestoreHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
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

	database.RestorePage(w, r, id)
}


func RenderTemplate(w http.ResponseWriter, tmpl string, p *database.WikiPage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(pages|revisions|preview)/(edit|save|view|preview|delete|restore|revisions|rollback|create)/([0-9]+)$")

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
