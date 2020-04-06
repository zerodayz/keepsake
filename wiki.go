// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bufio"
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
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

func (p *Page) save(datapath string) error {
	os.Mkdir("data", 0777)
	filename := datapath + p.Title + ".md"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (p *Page) validate(w http.ResponseWriter, r *http.Request) bool {
	p.Errors = make(map[string]string)
	if len(p.EditTitle) != 0 {
		match := validFilename.Match([]byte(p.EditTitle))
		if match == false {
			p.Errors["Title"] = "Please enter a valid title. Allowed charset: [a-zA-Z0-9_]"
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
		http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)
	}
	return len(p.Errors) == 0
}

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

func loadPage(datapath, title string) (*Page, error) {
	filename := datapath + title + ".md"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(datapath, title)

	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
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
		line = anchorReg.ReplaceAll(line, []byte(`<a href="/view/$1">$1</a>`))
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
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(datapath, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	editTitle := r.FormValue("title")
	p := &Page{Title: title, Body: []byte(body), EditTitle: editTitle}

	if p.validate(w, r) == false {
		renderTemplate(w, "edit", p)
		return
	}

	nlcr := regexp.MustCompile("\r\n")
	body = string(nlcr.ReplaceAllFunc([]byte(body), func(s []byte) []byte {
		return []byte("\n")
	}))

	p = &Page{Title: editTitle, Body: []byte(body)}
	err := p.save(datapath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if editTitle != title {
		os.Remove(datapath + title + ".md")
	}
	http.Redirect(w, r, "/view/"+editTitle, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := u.Query()
	searchKey := params.Get("q")
	var fileReg = regexp.MustCompile(`^[a-zA-Z0-9_]+\.md$`)
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
				buf.Write([]byte(`<h2><a href="/view/` +
					fileName[0] + `">` +
					fileName[0] + `</a></h2>`))
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
					buf.Write([]byte(content[showStart:start]))
					buf.Write([]byte(`<b>`))
					buf.Write([]byte(content[start:end]))
					buf.Write([]byte(`</b>`))
					buf.Write([]byte(content[end:showEnd]))
					buf.Write([]byte(`</pre></code>`))
				}
				buf.Write([]byte(`<br>`))
				buf.WriteByte('\n')
			}
		}
	}

	p := &Page{Title: searchKey, Body: []byte(buf.String())}
	p.DisplayBody = template.HTML(buf.String())
	p.Title = searchKey
	renderTemplate(w, "search", p)
}

var (
	tmplpath  = "tmpl/"
	datapath  = "data/"
	templates = template.Must(template.ParseFiles(
		tmplpath+"edit.html", tmplpath+"view.html", tmplpath+"search.html"))
	validFilename = regexp.MustCompile("^([a-zA-Z0-9_]+)$")
)

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9_]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/", rootHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
