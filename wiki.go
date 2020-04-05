// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"bufio"
	"bytes"
	"strings"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
	DisplayBody template.HTML
}

func (p *Page) save(datapath string) error {
	os.Mkdir("data", 0777)
	filename := datapath + p.Title + ".md"
	return ioutil.WriteFile(filename, p.Body, 0600)
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
		boldItalicReg = regexp.MustCompile(`\*\*\*(.*?)\*\*\*`)
		boldReg       = regexp.MustCompile(`\*\*(.*?)\*\*`)
		italicReg     = regexp.MustCompile(`\*(.*?)\*`)
		strikeReg     = regexp.MustCompile(`\~\~(.*?)\~\~`)
		underscoreReg = regexp.MustCompile(`__(.*?)__`)
		anchorReg     = regexp.MustCompile(`\[(.*?)\]`)
		anchorExtReg  = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
		escapeReg     = regexp.MustCompile(`^\>(\s|)`)
		blockquoteReg = regexp.MustCompile(`\&gt\;(.*?)$`)
		backtipReg    = regexp.MustCompile("`(.*?)`")
		h1Reg = regexp.MustCompile(`^#(\s|)(.*?)$`)
		h2Reg = regexp.MustCompile(`^##(\s|)(.*?)$`)
		h3Reg = regexp.MustCompile(`^###(\s|)(.*?)$`)
		h4Reg = regexp.MustCompile(`^####(\s|)(.*?)$`)
		h5Reg = regexp.MustCompile(`^#####(\s|)(.*?)$`)
		h6Reg = regexp.MustCompile(`^######(\s|)(.*?)$`)
		startBlock      bool = true
	)

	escapedBody := template.HTMLEscapeString(string(p.Body))
	buf := bytes.NewBuffer(nil)

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

	nlcr := regexp.MustCompile("\r\n")
	body = string(nlcr.ReplaceAllFunc([]byte(body), func(s []byte) []byte {
		return []byte("\n")
	}))

	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(datapath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var (
	tmplpath = "tmpl/"
	datapath = "data/"
	templates = template.Must(template.ParseFiles(tmplpath+"edit.html", tmplpath+"view.html"))
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
	http.HandleFunc("/", rootHandler)
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}