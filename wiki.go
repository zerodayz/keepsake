// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"gowiki/pages"
	"log"
	"net/http"
	"regexp"
)

var validPath = regexp.MustCompile("^/(pages)/(edit|save|view)/([a-z0-9_]+)$")

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pages/view/home", http.StatusFound)
}

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

func main() {
	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.HandleFunc("/pages/view/", MakeHandler(pages.ViewHandler))
	http.HandleFunc("/pages/edit/", MakeHandler(pages.EditHandler))
	http.HandleFunc("/pages/save/", MakeHandler(pages.SaveHandler))
	http.HandleFunc("/pages/search/", pages.SearchHandler)
	http.HandleFunc("/", RootHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
