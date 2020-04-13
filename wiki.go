// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"gowiki/pages"
	"gowiki/users"
	"gowiki/database"
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pages/view/home", http.StatusFound)
}

func main() {
	database.InitializeDatabase()

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.HandleFunc("/pages/view/", pages.MakeHandler(pages.ViewHandler))
	http.HandleFunc("/pages/edit/", pages.MakeHandler(pages.EditHandler))
	http.HandleFunc("/pages/save/", pages.MakeHandler(pages.SaveHandler))
	http.HandleFunc("/pages/search/", pages.SearchHandler)
	http.HandleFunc("/users/login/", users.LoginHandler)
	http.HandleFunc("/users/logout/", users.LogoutHandler)
	http.HandleFunc("/users/create/", users.CreateUserHandler)
	http.HandleFunc("/", RootHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
