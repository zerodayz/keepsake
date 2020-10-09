package pages

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/zerodayz/keepsake/database"
	"github.com/zerodayz/keepsake/helpers"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	text "text/template"
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
	return &database.WikiPage{Title: s.Title, Body: s.Content, Tags: s.Tags, Deleted: s.Deleted, InternalId: InternalId, CreatedBy: s.CreatedBy, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated, Liked: s.Liked}, nil
}

func LoadPreviewPage(w http.ResponseWriter, r *http.Request, InternalId int) (*database.WikiPage, error) {
	s := database.ShowPreviewPage(w, r, InternalId)
	return &database.WikiPage{Title: s.Title, Body: s.Content, Tags: s.Tags, InternalId: InternalId, CreatedBy: s.CreatedBy, LastModified: s.LastModified, LastModifiedBy: s.LastModifiedBy, DateCreated: s.DateCreated}, nil
}

func LoadRevisionPage(w http.ResponseWriter, r *http.Request, InternalId int) (*database.WikiPageRevision, *database.WikiPage) {
	wpr, wp := database.ShowRevisionPage(w, r, InternalId)
	return &database.WikiPageRevision{Title: wpr.Title, WikiPageId: wpr.WikiPageId, Tags: wpr.Tags, RevisionId: wpr.RevisionId, Content: wpr.Content, InternalId: InternalId, CreatedBy: wpr.CreatedBy, LastModified: wpr.LastModified, LastModifiedBy: wpr.LastModifiedBy, DateCreated: wpr.DateCreated},
		wp
}

// Handlers

func ViewRawHandler(w http.ResponseWriter, r *http.Request) {
	var validPathRaw = regexp.MustCompile("^/pages/view/raw/([0-9]+)$")
	m := validPathRaw.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	InternalId := m[1]
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}
	s, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	data := "# " + s.Title + "\n" + s.Body
	tmpl, err := text.New("/lib/pages/raw.md").Delims("{%i do not know what to type%{", "}%hope this solves my problem%}").Parse(data)
	w.Header().Set("content-type", "text/markdown")

	err = tmpl.Execute(w, s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SearchRawHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var searchTrimNum = 100
	params := u.Query()
	searchKey := params.Get("q")
	searchTrim := params.Get("s")
	if len(searchTrim) != 0 {
		searchTrimNum, _ = strconv.Atoi(searchTrim)
	}
	var searchQuery = regexp.MustCompile(`(?i)` + searchKey)
	buf := bytes.NewBuffer(nil)

	if len(searchKey) == 0 {
		buf.Write([]byte(`Please search for something else than empty.`))
	} else {
		s := database.SearchWikiPages(w, r, searchKey)
		for _, f := range s {
			var contentLength = len(f.Content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var indexes = searchQuery.FindAllIndex([]byte(f.Content), -1)

			var originalTags []string
			for i, s := range f.Tags {
				originalTags = append(originalTags, s)
				space := strings.Split(s, " ")
				if len(space) >= 2 {
					f.Tags[i] = strings.ReplaceAll(s, " ", "-")
				}
			}
			if len(indexes) != 0 {
				buf.Write([]byte("\n" + ` ================== ` + "\n"))
				if f.Deleted == 1 {
					buf.Write([]byte(`>> Matched Article: ` + f.Title + ` (deleted)` +
						"\n" + `For more information please curl endpoint at /pages/view/raw/` + strconv.Itoa(f.InternalId)))
				} else {
					buf.Write([]byte(`>> Matched Article: ` + f.Title +
						"\n" + `For more information please curl endpoint at /pages/view/raw/` + strconv.Itoa(f.InternalId)))
				}
				buf.Write([]byte("\n" + ` ================== ` + "\n"))
				buf.Write([]byte("\n" + ` @@@@@@@@@@@@@@@@@@ ` + "\n"))
				for _, k := range indexes {
					var start = k[0]
					var end = k[1]

					var showStart = max(start-searchTrimNum, 0)
					var showEnd = min(end+searchTrimNum, contentLength-1)

					for !utf8.RuneStart(f.Content[showStart]) {
						showStart = max(showStart-1, 0)
					}
					for !utf8.RuneStart(f.Content[showEnd]) {
						showEnd = min(showEnd-1, contentLength)
					}
					buf.WriteString(f.Content[showStart:start])
					buf.WriteString(f.Content[start:end])
					if (end - 1) != showEnd {
						buf.WriteString(f.Content[end:showEnd])
					}
				}
				buf.Write([]byte("\n" + ` @@@@@@@@@@@@@@@@@@ ` + "\n"))
			} else {
				buf.Write([]byte("\n" + ` ================== ` + "\n"))
				buf.Write([]byte(`>> Matched Title: ` + f.Title +
					"\n" + `For more please visit /pages/view/` + strconv.Itoa(f.InternalId)))
				buf.Write([]byte("\n" + ` ================== ` + "\n"))
			}
		}
	}
	data := buf.String()
	tmpl, err := text.New("/lib/pages/search.md").Parse(data)
	w.Header().Set("content-type", "text/markdown")

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("solarized-dark256"),
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
	if username == "Unauthorized" {
		s.Liked = 2
		s.Repair = 2
	} else {
		s.Liked = database.GetLikeForPagePerUser(w, r, id, username)
		s.Repair = database.GetRepairsForPage(w, r, id)
	}

	bufComments.Write([]byte(`<div>There are ` + strconv.Itoa(len(comments)) + ` comment(s).</div>`))
	for _, f := range comments {
		bufComments.Write([]byte(`
				<div class="found">Comment by ` + f.CreatedBy + ` on ` + f.DateCreated + `
				<label for="search-content" class="search-collapsible">` + f.Title + `</label>
				<div id="search-content" class="comment-content">
				<pre><code>` + f.Body + `</code></pre></div></div>`))
	}
	s.DisplayComment = template.HTML(bufComments.String())

	err = t.ExecuteTemplate(w, "view.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := ioutil.TempDir("", "keepsake-latest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var files []string
	wikiPages := database.DownloadAllPages(w, r)
	for _, f := range wikiPages {
		fileName := strings.ReplaceAll(f.Title, " ", "_") + ".md"
		fileName = strings.ReplaceAll(fileName, "/", "_")
		files = append(files, dir + "/" + fileName)
		data := "# " + f.Title + "\n" + f.Content
		file := filepath.Join(dir, fileName)
		if err := ioutil.WriteFile(file, []byte(data), 0666);
			err != nil {
			log.Fatal(err)
		}
	}
	err = helpers.CreateTarball("lib/keepsake-latest.tar.gz", files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/tar+gzip")
	http.Redirect(w, r, "/lib/keepsake-latest.tar.gz", http.StatusFound)
	return
}

func ListRepairsHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "repairs.html"))

	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}

	bufUpVoted := bytes.NewBuffer(nil)
	bufUpVoted.Write([]byte(`<div class="container-d">
			<div class="header-text left-d"><h1>View Needs Improvement</h1></div>
    <form id="searchForm" action="/pages/search" method="GET">
        <div class="control-group search-container right-d">
            <div class="controls">
              <input type="search" class="search-input" id="inputQuery" name="q" placeholder="Search" value="">
            </div>
            <div class="control-group">
                <div class="controls">
                    <input class="navbar-search-button" id="submit" type="submit" value="Search">
                </div>
            </div>
        </div>
    </form>
	</div>`))

	needsImprovementPages := database.LoadNeedsImprovement(w, r)

	if len(needsImprovementPages) == 0 {
		bufUpVoted.Write([]byte(`There are no pages marked as Needs Improvement. Everything is good :-)`))
	} else {
		existingCategories := database.FetchCategories(w, r)
		if len(existingCategories) == 0 {
			bufUpVoted.Write([]byte(`There are no categories yet.`))
		} else {
			bufUpVoted.Write([]byte(`<div id="items"></div>`))
			for _, f := range existingCategories {
				// Fix for the categories with space.
				space := strings.Split(f.Name, " ")
				if len(space) >= 2 {
					value := strings.ReplaceAll(f.Name, " ", "-")
					bufUpVoted.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + value + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				} else {
					bufUpVoted.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				}
			}
		}
		bufUpVoted.Write([]byte(`
		<table id="view-all-table">
		`))
		for _, f := range needsImprovementPages {
			var originalTags []string
			for i, s := range f.Tags {
				originalTags = append(originalTags, s)
				space := strings.Split(s, " ")
				if len(space) >= 2 {
					f.Tags[i] = strings.ReplaceAll(s, " ", "-")
				}
			}
			categoriesName := strings.Join(originalTags, " ")
			categories := strings.Join(f.Tags, " ")

			if len(categories) == 0 {
				bufUpVoted.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: None | Requested by ` + f.CreatedBy + `</td></tr>`))
			} else {
				bufUpVoted.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: ` + categoriesName + ` | Requested by ` + f.CreatedBy + `</td></tr>`))
			}
		}
		bufUpVoted.Write([]byte(`</table>`))
	}
	p.DisplayBody = template.HTML(bufUpVoted.String())

	err := t.ExecuteTemplate(w, "repairs.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StarHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "stars.html"))

	username := ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}

	bufUpVoted := bytes.NewBuffer(nil)
	bufUpVoted.Write([]byte(`<div class="container-d">
			<div class="header-text left-d"><h1>View My Voted</h1></div>
    <form id="searchForm" action="/pages/search" method="GET">
        <div class="control-group search-container right-d">
            <div class="controls">
              <input type="search" class="search-input" id="inputQuery" name="q" placeholder="Search" value="">
            </div>
            <div class="control-group">
                <div class="controls">
                    <input class="navbar-search-button" id="submit" type="submit" value="Search">
                </div>
            </div>
        </div>
    </form>
	</div>`))

	upVotedPages := database.LoadMyVoted(w, r, username)

	if len(upVotedPages) == 0 {
		bufUpVoted.Write([]byte(`There are no wiki pages with your votes :-(`))
	} else {
		existingCategories := database.FetchCategories(w, r)
		if len(existingCategories) == 0 {
			bufUpVoted.Write([]byte(`There are no categories yet.`))
		} else {
			bufUpVoted.Write([]byte(`<div id="items"></div>`))
			for _, f := range existingCategories {
				// Fix for the categories with space.
				space := strings.Split(f.Name, " ")
				if len(space) >= 2 {
					value := strings.ReplaceAll(f.Name, " ", "-")
					bufUpVoted.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + value + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				} else {
					bufUpVoted.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				}
			}
		}
		bufUpVoted.Write([]byte(`
		<table id="view-all-table">
		`))
		for _, f := range upVotedPages {
			var originalTags []string
			for i, s := range f.Tags {
				originalTags = append(originalTags, s)
				space := strings.Split(s, " ")
				if len(space) >= 2 {
					f.Tags[i] = strings.ReplaceAll(s, " ", "-")
				}
			}
			categoriesName := strings.Join(originalTags, " ")
			categories := strings.Join(f.Tags, " ")

			if len(categories) == 0 {
				bufUpVoted.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: None</td></tr>`))
			} else {
				bufUpVoted.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: ` + categoriesName + `</td></tr>`))
			}
		}
		bufUpVoted.Write([]byte(`</table>`))
	}
	p.DisplayBody = template.HTML(bufUpVoted.String())

	err := t.ExecuteTemplate(w, "stars.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "list.html"))

	username := ReadCookie(w, r)
	buf := bytes.NewBuffer(nil)
	wikiPages := database.LoadAllPages(w, r)
	buf.Write([]byte(`<div class="container-d">
			<div class="header-text left-d"><h1>View All <a href="/pages/download/"><img src="/lib/icons/get_app-24px.svg"></a></h1></div>
    <form id="searchForm" action="/pages/search" method="GET">
        <div class="control-group search-container right-d">
            <div class="controls">
              <input type="search" class="search-input" id="inputQuery" name="q" placeholder="Search" value="">
            </div>
            <div class="control-group">
                <div class="controls">
                    <input class="navbar-search-button" id="submit" type="submit" value="Search">
                </div>
            </div>
        </div>
    </form>
	</div>`))
	if len(wikiPages) == 0 {
		buf.Write([]byte(`There are no wiki pages.`))
	} else {
		existingCategories := database.FetchCategories(w, r)
		if len(existingCategories) == 0 {
			buf.Write([]byte(`There are no categories yet.`))
		} else {
			buf.Write([]byte(`<div id="items"></div>`))
			for _, f := range existingCategories {
				// Fix for the categories with space.
				space := strings.Split(f.Name, " ")
				if len(space) >= 2 {
					value := strings.ReplaceAll(f.Name, " ", "-")
					buf.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + value + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				} else {
					buf.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				}
			}
		}
		buf.Write([]byte(`
		<table id="view-all-table">
		`))
		for _, f := range wikiPages {
			var originalTags []string
			for i, s := range f.Tags {
				originalTags = append(originalTags, s)
				space := strings.Split(s, " ")
				if len(space) >= 2 {
					f.Tags[i] = strings.ReplaceAll(s, " ", "-")
				}
			}
			categoriesName := strings.Join(originalTags, " ")
			categories := strings.Join(f.Tags, " ")

			if len(categories) == 0 {
				buf.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: None</td></tr>`))
			} else {
				buf.Write([]byte(`<tr><td class="dashboard category `+ categories + `"><a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br>Categories: ` + categoriesName + `</td></tr>`))
			}
		}
		buf.Write([]byte(`</table>`))
	}
	p.DisplayBody = template.HTML(buf.String())

	p.UserLoggedIn = username
	err := t.ExecuteTemplate(w, "list.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LikeHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
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
	database.LikePage(w, r, id, username)
}

func RepairHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	c := database.Comment{}
	username := ReadCookie(w, r)
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
		c.Title = "Needs Improvement Requested: " + r.PostFormValue("comment_title_needs_improvement")
		c.Body = r.PostFormValue("comment_message_needs_improvement")
		date := time.Now().UTC()
		c.DateCreated = date.Format("20060102150405")
		database.RepairPageAndComment(w, r, id, c)
		return
	}
	database.RepairPage(w, r, id, username)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	p := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "dashboard.html"))

	username := ReadCookie(w, r)
	dateYesterday := time.Now().AddDate(0, 0, -1).UTC()

	bufTodayArticles := bytes.NewBuffer(nil)
	wikiPagesToday := database.LoadAllPagesToday(w, r)
	bufTodayArticles.Write([]byte(`
			<div class="container-d">
                <div class="header-text left-d"><h1>Today's Articles</h1></div>
                <form id="searchForm" action="/pages/search" method="GET">
                    <div class="control-group search-container right-d">
                        <div class="controls">
                            <input type="search" class="search-input" id="inputQuery" name="q" placeholder="Search" value="">
                        </div>
                        <div class="control-group">
                            <div class="controls">
                                <input class="navbar-search-button" id="submit" type="submit" value="Search">
                            </div>
                        </div>
                    </div>
                </form>
            </div>`))

	if len(wikiPagesToday) == 0 {
		bufTodayArticles.Write([]byte(`There are no Wiki pages created today yet :-(`))
	} else {
		for _, f := range wikiPagesToday {
			// 2020-08-02 23:44:28
			dateCreated := time.Now()

			dateCreated, err := time.Parse("2006-01-02 15:04:05", f.DateCreated)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if f.LastModifiedBy == "" {
				if dateCreated.After(dateYesterday) {
					bufTodayArticles.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/fiber_new-24px.svg" alt="New post!" title="New post!"/> | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						`</div>`))
				}
			} else {
				dateModified, err := time.Parse("2006-01-02 15:04:05", f.LastModified)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				if dateCreated.After(dateYesterday) && dateModified.After(dateYesterday) {
					bufTodayArticles.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/fiber_new-24px.svg" alt="New post!" title="New post!"/> <img src="/lib/icons/new_releases-24px.svg" alt="New update!" title="New update!"/> | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				}
			}
		}
	}

	bufComment := bytes.NewBuffer(nil)
	wikiPagesTop10Commented := database.Top10Commented(w, r)
	bufComment.Write([]byte(`<div class="header-text"><h1>Last 10 Discussed</h1></div>`))
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
				bufComment.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/comment-24px.svg" alt="New comment!"/> | Comments: ` + strconv.Itoa(len(comments)) + ` | Last commented on ` + f.DateCreated + ` by ` + f.CreatedBy + `</div>`))
			} else {
				bufComment.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> Comments: ` + strconv.Itoa(len(comments)) + ` | Commented on ` + f.DateCreated + ` by ` + f.CreatedBy + `</div>`))
			}
		}
	}

	buf := bytes.NewBuffer(nil)
	wikiPages := database.LoadPageLast25(w, r)
	buf.Write([]byte(`<div class="header-text-n"><h1>Last 25 Updated</h1></div>`))
	if len(wikiPages) == 0 {
		buf.Write([]byte(`There are no wiki pages.`))
	} else {
		for _, f := range wikiPages {
			// 2020-08-02 23:44:28
			dateCreated := time.Now()

			dateCreated, err := time.Parse("2006-01-02 15:04:05", f.DateCreated)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if f.LastModifiedBy == "" {
				if dateCreated.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/fiber_new-24px.svg" alt="New post!" title="New post!"/> | Created on ` + f.DateCreated + ` by ` + f.CreatedBy +
						`</div>`))
				} else {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> Created on ` + f.DateCreated + ` by ` + f.CreatedBy + `</div>`))
				}
			} else {
				dateModified, err := time.Parse("2006-01-02 15:04:05", f.LastModified)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				if dateCreated.After(dateYesterday) && dateModified.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/fiber_new-24px.svg" alt="New post!" title="New post!"/> <img src="/lib/icons/new_releases-24px.svg" alt="New update!" title="New update!"/> | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				} else if dateModified.After(dateYesterday) {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> <img src="/lib/icons/new_releases-24px.svg" alt="New update!" title="New update!"/> | Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				} else {
					buf.Write([]byte(`<div class="dashboard"> <a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a> <br> Modified on ` + f.LastModified + ` by ` + f.LastModifiedBy + `.</div>`))
				}
			}
		}
	}
	bufUpVoted := bytes.NewBuffer(nil)
	bufUpVoted.Write([]byte(`<div class="header-text-n"><h1>5 Most Voted</h1></div>`))

	upVotedPages := database.LoadTop5Voted(w, r)

	if len(upVotedPages) == 0 {
		bufUpVoted.Write([]byte(`There are no wiki pages with votes.`))
	} else {
		for _, f := range upVotedPages {
			bufUpVoted.Write([]byte(`<div class="dashboard">
			<a class="dashboard-title" href="/pages/view/` + strconv.Itoa(f.InternalId) + `">` + f.Title + `</a><br>Votes: ` + strconv.Itoa(f.Liked) + `</div>`))
		}
	}
	p.DisplayUpVoted = template.HTML(bufUpVoted.String())
	p.DisplayToday = template.HTML(bufTodayArticles.String())
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
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("solarized-dark256"),
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
		struct{ WikiPageRevision, WikiPage interface{} }{wpr, wp})
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
	bufCategories := bytes.NewBuffer(nil)
	existingCategories := database.FetchCategories(w, r)
	if len(existingCategories) == 0 {
		bufCategories.Write([]byte(`There are no categories yet.`))
	} else {
		var matched bool
		for _, f := range existingCategories {
			matched = false
			for _, tag := range s.Tags {
				if f.Name == tag {
					bufCategories.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox" checked>` + f.Name + `<span class="checkmark"></span></label></div>`))
					matched = true
				}
			}
			if matched != true {
				bufCategories.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
			}
		}
	}
	s.DisplayComment = template.HTML(bufCategories.String())

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
		s.Tags = r.Form["tags"]
		date := time.Now().UTC()
		s.DateCreated = date.Format("20060102150405")
		s.InternalId = database.CreatePreviewPage(w, r, s)
		http.Redirect(w, r, "/preview/view/"+strconv.Itoa(s.InternalId), http.StatusFound)

	}
	bufCategories := bytes.NewBuffer(nil)
	existingCategories := database.FetchCategories(w, r)
	if len(existingCategories) == 0 {
		bufCategories.Write([]byte(`There are no categories yet.`))
	} else {
		for _, f := range existingCategories {
			bufCategories.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
		}
	}
	s.DisplayComment = template.HTML(bufCategories.String())

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
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("solarized-dark256"),
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

func UploadFile(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(templatePath + "upload.html"))
	f := database.File{}

	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file-upload-field")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	tempFile, err := ioutil.TempFile("uploads", "upload-*")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}
	tempFile.Write(fileBytes)
	f.Name = tempFile.Name()

	err = t.ExecuteTemplate(w, "upload.html", f)
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
		s.Tags = r.Form["tags"]
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
	} else {
		buf.Write([]byte(`<div id="items"></div>`))
		existingCategories := database.FetchCategories(w, r)
		if len(existingCategories) == 0 {
			buf.Write([]byte(`There are no categories yet.`))
		} else {
			for _, f := range existingCategories {
				// Fix for the categories with space.
				space := strings.Split(f.Name, " ")
				if len(space) >= 2 {
					value := strings.ReplaceAll(f.Name, " ", "-")
					buf.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + value + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				} else {
					buf.Write([]byte(`<div class="categories"><label class="checkbox"><input name="tags" value="` + f.Name + `" type="checkbox">` + f.Name + `<span class="checkmark"></span></label></div>`))
				}			}
		}
		s := database.SearchWikiPages(w, r, searchKey)
		for _, f := range s {
			var contentLength = len(f.Content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var categories string
			var indexes = searchQuery.FindAllIndex([]byte(f.Content), -1)

			var originalTags []string
			for i, s := range f.Tags {
				originalTags = append(originalTags, s)
				space := strings.Split(s, " ")
				if len(space) >= 2 {
					f.Tags[i] = strings.ReplaceAll(s, " ", "-")
				}
			}
			var categoriesName string
			if len(strings.Join(f.Tags, " ")) >= 1 {
				categoriesName = strings.Join(originalTags, " ")
				categories = strings.Join(f.Tags, " ")
			} else {
				categories = "None"
				categoriesName = "None"
			}
			if len(indexes) != 0 {
				var occurrences = strconv.Itoa(len(indexes))
				if f.Deleted == 0 {
					buf.Write([]byte(`
					<div class="category ` + categories + `">
					<div class="found">Found ` + occurrences + ` occurrences.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` +
						` | Categories: ` + categoriesName +
						`<label for="search-content" class="search-collapsible">
					` + f.Title + `</label>
					<div id="search-content" class="search-content">`))
				} else {
					buf.Write([]byte(`
					<div class="category ` + categories + `">
					<div class="found">Found ` + occurrences + ` occurrences.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a>` +
						` | Categories: ` + categoriesName +
						`<label for="search-content" class="search-collapsible">
					` + f.Title + ` (deleted)</label>
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
				buf.Write([]byte(`</div></div></div>`))
				buf.WriteByte('\n')
			} else {
				if f.Deleted == 0 {
					buf.Write([]byte(`
					<div class="category ` + categories + `">
					<div class="found">Matched title of the wiki page.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` +
						` | Categories: ` + categoriesName +
						`<label for="search-content" class="search-no-collapsible">
					` + f.Title + `</label>`))
				} else {
					buf.Write([]byte(`
					<div class="category ` + categories + `">
					<div class="found">Matched title of the wiki page.
					<a href="/pages/view/` + strconv.Itoa(f.InternalId) + `">Visit Page</a> ` +
						` | Categories: ` + categoriesName +
						`<label for="search-content" class="search-no-collapsible">
					` + f.Title + ` (deleted)</label>`))
				}
				buf.Write([]byte(`</div></div>`))
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
		if len(f.Title) == 0 {
			f.Title = "Missing title"
		}
		if len(wikiRevisionPages) == f.RevisionId {
			buf.Write([]byte(`<b>` + f.Title + `</b><br><a href="/revisions/view/` + strconv.Itoa(f.InternalId) + `" target="_parent">Latest version </a> | ` + `Last Modified by ` +
				f.LastModifiedBy + ` on ` + f.LastModified))
			buf.Write([]byte(`<br>`))
		} else {
			buf.Write([]byte(`<b>` + f.Title + `</b><br><a href="/revisions/view/` + strconv.Itoa(f.InternalId) + `" target="_parent">Revision ` + strconv.Itoa(f.RevisionId) + `</a> | ` + `Last Modified by ` +
				f.LastModifiedBy + ` on ` + f.LastModified))
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
		buf.Write([]byte(`<div class="dashboard"> <a href="/pages/view/` + strconv.Itoa(f.InternalId) + `" class="dashboard-title-recycle-bin"> ` + f.Title + `</a><br> <a class="link-recycle-bin" href="/pages/restore/` + strconv.Itoa(f.InternalId) + `">Remove from Bin</a></div>`))
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

var validPath = regexp.MustCompile("^/(pages|revisions|preview)/(like|repair|unlike|edit|save|view|preview|delete|restore|revisions|rollback|create)/([0-9]+)$")

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
