// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"github.com/zerodayz/gowiki/database"
	"github.com/zerodayz/gowiki/pages"
	"github.com/zerodayz/gowiki/users"
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pages/view/1", http.StatusFound)
}

func main() {
	database.InitializeDatabase()

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.HandleFunc("/revisions/view/", pages.MakeHandler(pages.RevisionsViewHandler))
	http.HandleFunc("/pages/view/", pages.MakeHandler(pages.ViewHandler))
	http.HandleFunc("/pages/revisions/", pages.MakeHandler(pages.RevisionsHandler))
	http.HandleFunc("/pages/edit/", pages.MakeHandler(pages.EditHandler))
	http.HandleFunc("/pages/delete/", pages.MakeHandler(pages.DeleteHandler))
	http.HandleFunc("/pages/create/", pages.CreateHandler)
	http.HandleFunc("/pages/save/", pages.MakeHandler(pages.SaveHandler))
	http.HandleFunc("/pages/trash/", pages.RecycleBinHandler)
	http.HandleFunc("/pages/restore/", pages.RestoreHandler)
	//http.HandleFunc("/pages/search/", pages.SearchHandler)
	http.HandleFunc("/users/login/", users.LoginHandler)
	http.HandleFunc("/users/logout/", users.LogoutHandler)
	http.HandleFunc("/users/create/", users.CreateUserHandler)
	http.HandleFunc("/", RootHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
