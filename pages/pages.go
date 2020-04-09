package pages

import (
	"bufio"
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Page struct {
	Title       string
	EditTitle   string
	Body        []byte
	DisplayBody template.HTML
	Errors      map[string]string
}

var (
	tmplpath  = "tmpl/"
	datapath  = "data/pages/"
	templates = template.Must(template.ParseFiles(
		tmplpath+"edit.html", tmplpath+"view.html", tmplpath+"search.html"))
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

func (p *Page) Save(datapath string) error {
	os.MkdirAll(datapath, 0777)
	filename := datapath + p.Title + ".md"
	return ioutil.WriteFile(filename, p.Body, 0600)
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

func LoadPage(datapath, title string) (*Page, error) {
	filename := datapath + title + ".md"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// Handlers

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := LoadPage(datapath, title)

	if err != nil {
		http.Redirect(w, r, "/pages/edit/"+title, http.StatusFound)
		return
	}

	var (
		boldItalicReg      = regexp.MustCompile(`\*\*\*(.*?)\*\*\*`)
		boldReg            = regexp.MustCompile(`\*\*(.*?)\*\*`)
		italicReg          = regexp.MustCompile(`\*(.*?)\*`)
		strikeReg          = regexp.MustCompile(`\~\~(.*?)\~\~`)
		underscoreReg      = regexp.MustCompile(`__(.*?)__`)
		anchorReg          = regexp.MustCompile(`\[(.*?)\]`)
		anchorExtReg       = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
		escapeReg          = regexp.MustCompile(`^\>(\s|)`)
		blockquoteReg      = regexp.MustCompile(`\&gt\;(.*?)$`)
		backtipReg         = regexp.MustCompile("`(.*?)`")
		h1Reg              = regexp.MustCompile(`^#(\s|)(.*?)$`)
		h2Reg              = regexp.MustCompile(`^##(\s|)(.*?)$`)
		h3Reg              = regexp.MustCompile(`^###(\s|)(.*?)$`)
		h4Reg              = regexp.MustCompile(`^####(\s|)(.*?)$`)
		h5Reg              = regexp.MustCompile(`^#####(\s|)(.*?)$`)
		h6Reg              = regexp.MustCompile(`^######(\s|)(.*?)$`)
		startBlock    bool = true
		startToc      bool = true
	)

	escapedBody := template.HTMLEscapeString(string(p.Body))
	buf := bytes.NewBuffer(nil)

	tocscanner := bufio.NewScanner(strings.NewReader(escapedBody))
	for tocscanner.Scan() {

		line := tocscanner.Bytes()
		if len(line) != 0 {
			if string(line) == "----" {
				if startBlock == true {
					startBlock = false
				} else {
					startBlock = true
				}
				continue
			}
			if startBlock == false {
				continue
			}
			if line[0] == '#' {
				if startToc == true {
					buf.Write([]byte(`
					<input id="collapsible" class="toggle" type="checkbox">
					<label for="collapsible" class="lbl-toggle">Table of Content</label>
					<div class="collapsible-content">`))
					startToc = false
				}
				count := bytes.Count(line, []byte(`#`))
				switch count {
				case 1:
					line = h1Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h1">$2</a>`))
				case 2:
					line = h2Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h2">$2</a>`))
				case 3:
					line = h3Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h3">$2</a>`))
				case 4:
					line = h4Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h4">$2</a>`))
				case 5:
					line = h5Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h5">$2</a>`))
				case 6:
					line = h6Reg.ReplaceAll(line, []byte(`<a href="#$2" class="h6">$2</a>`))
				}
				startToc = false
				buf.Write(line)
				buf.Write([]byte(`<br>`))
				buf.WriteByte('\n')
			}
		}
	}
	if startToc == false {
		buf.Write([]byte(`</div>`))
		buf.Write([]byte(`<br>`))
		buf.WriteByte('\n')
	}

	scanner := bufio.NewScanner(strings.NewReader(escapedBody))
	for scanner.Scan() {

		line := scanner.Bytes()
		if len(line) == 0 {
			buf.Write([]byte(`<br>`))
			buf.WriteByte('\n')
			continue
		}
		if string(line) == "----" {
			if startBlock == true {
				buf.Write([]byte(`<pre><code>`))
				startBlock = false
			} else {
				buf.Write([]byte(`</code></pre>`))
				startBlock = true
			}
			continue
		}
		if startBlock == false {
			buf.Write(line)
			buf.WriteByte('\n')
			continue
		}
		line = boldItalicReg.ReplaceAll(line, []byte(`<b><i>$1</i></b>`))
		line = boldReg.ReplaceAll(line, []byte(`<b>$1</b>`))
		line = italicReg.ReplaceAll(line, []byte(`<i>$1</i>`))
		line = strikeReg.ReplaceAll(line, []byte(`<s>$1</s>`))
		line = underscoreReg.ReplaceAll(line, []byte(`<u>$1</u>`))
		line = anchorExtReg.ReplaceAll(line, []byte(`<a href="$2">$1</a>`))
		line = anchorReg.ReplaceAll(line, []byte(`<a href="/pages/view/$1">$1</a>`))
		line = escapeReg.ReplaceAll(line, []byte(`&gt;`))
		line = blockquoteReg.ReplaceAll(line, []byte(`<blockquote>$1</blockquote>`))
		line = backtipReg.ReplaceAll(line, []byte(`<code>$1</code>`))
		if line[0] == '#' {
			count := bytes.Count(line, []byte(`#`))
			switch count {
			case 1:
				line = h1Reg.ReplaceAll(line, []byte(`<h1 id="$2">$2</h1>`))
			case 2:
				line = h2Reg.ReplaceAll(line, []byte(`<h2 id="$2">$2</h2>`))
			case 3:
				line = h3Reg.ReplaceAll(line, []byte(`<h3 id="$2">$2</h3>`))
			case 4:
				line = h4Reg.ReplaceAll(line, []byte(`<h4 id="$2">$2</h4>`))
			case 5:
				line = h5Reg.ReplaceAll(line, []byte(`<h5 id="$2">$2</h5>`))
			case 6:
				line = h6Reg.ReplaceAll(line, []byte(`<h6 id="$2">$2</h6>`))
			}
		}
		buf.Write(line)
		buf.Write([]byte(`<br>`))
		buf.WriteByte('\n')
	}

	p.DisplayBody = template.HTML(buf.String())
	RenderTemplate(w, "view", p)
}

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := LoadPage(datapath, title)
	if err != nil {
		p = &Page{Title: title}
	}
	RenderTemplate(w, "edit", p)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	editTitle := r.FormValue("title")
	p := &Page{Title: title, Body: []byte(body), EditTitle: editTitle}

	if p.Validate(w, r) == false {
		RenderTemplate(w, "edit", p)
		return
	}

	nlcr := regexp.MustCompile("\r\n")
	body = string(nlcr.ReplaceAllFunc([]byte(body), func(s []byte) []byte {
		return []byte("\n")
	}))

	if editTitle != title {
		err := os.Rename(datapath+title+".md", datapath+editTitle+".md")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err := p.Save(datapath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/pages/view/"+editTitle, http.StatusFound)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := u.Query()
	searchKey := params.Get("q")
	var fileReg = regexp.MustCompile(`^[a-z0-9_]+\.md$`)
	var searchQuery = regexp.MustCompile(searchKey)

	buf := bytes.NewBuffer(nil)

	files, err := ioutil.ReadDir(datapath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(searchKey) == 0 {
		return
	}
	buf.Write([]byte(`<div id="items"></div>`))
	for _, f := range files {
		if fileReg.MatchString(f.Name()) {
			content, err := ioutil.ReadFile(datapath + f.Name())
			var contentLenght = len(content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var indexes = searchQuery.FindAllIndex(content, -1)
			if len(indexes) != 0 {
				fileName := strings.Split(f.Name(), ".")
				buf.Write([]byte(`<label for="search-content" class="search-collapsible">
				` + fileName[0] + `</label>
				<div id="search-content" class="search-content">
				<a href="/pages/view/` + fileName[0] + `"><img src="/lib/icons/visit-24px.svg"></a>
				<a href="/pages/edit/` + fileName[0] + `"><img src="/lib/icons/edit-outline-24px.svg"></a>`))
				for _, k := range indexes {
					var start = k[0]
					var end = k[1]

					var showStart = max(start-100, 0)
					var showEnd = min(end+100, contentLenght-1)

					for !utf8.RuneStart(content[showStart]) {
						showStart = max(showStart-1, 0)
					}
					for !utf8.RuneStart(content[showEnd]) {
						showEnd = min(showEnd-1, contentLenght)
					}
					buf.Write([]byte(`<pre><code>`))
					buf.WriteString(template.HTMLEscapeString(string(content[showStart:start])))
					buf.Write([]byte(`<b>`))
					buf.WriteString(template.HTMLEscapeString(string(content[start:end])))
					buf.Write([]byte(`</b>`))
					if (end - 1) != showEnd {
						buf.WriteString(template.HTMLEscapeString(string(content[end:showEnd])))
					}
					buf.Write([]byte(`</code></pre>`))
				}
				buf.Write([]byte(`</div>`))
				buf.WriteByte('\n')
			}
		}
	}

	p := &Page{Title: searchKey, Body: []byte(buf.String())}
	p.DisplayBody = template.HTML(buf.String())
	p.Title = searchKey
	RenderTemplate(w, "search", p)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
