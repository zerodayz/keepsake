// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"github.com/zerodayz/gowiki/database"
	"github.com/zerodayz/gowiki/pages"
	"github.com/zerodayz/gowiki/categories"
	"github.com/zerodayz/gowiki/users"
	"github.com/zerodayz/gowiki/comments"
	"github.com/zerodayz/gowiki/tickets"
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func main() {
	database.InitializeDatabase()

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.HandleFunc("/revisions/view/", pages.MakeHandler(pages.RevisionsViewHandler))
	http.HandleFunc("/revisions/rollback/", pages.MakeHandler(pages.RevisionRollbackHandler))
	http.HandleFunc("/preview/view/", pages.MakeHandler(pages.PreviewHandler))
	http.HandleFunc("/preview/create/", pages.MakeHandler(pages.PreviewCreateHandler))

	http.HandleFunc("/pages/view/", pages.MakeHandler(pages.ViewHandler))
	http.HandleFunc("/pages/revisions/", pages.MakeHandler(pages.RevisionsHandler))
	http.HandleFunc("/pages/edit/", pages.MakeHandler(pages.EditHandler))
	http.HandleFunc("/pages/delete/", pages.MakeHandler(pages.DeleteHandler))
	http.HandleFunc("/pages/create/", pages.CreateHandler)
	http.HandleFunc("/pages/save/", pages.MakeHandler(pages.SaveHandler))
	http.HandleFunc("/pages/trash/", pages.RecycleBinHandler)
	http.HandleFunc("/pages/restore/", pages.MakeHandler(pages.RestoreHandler))
	http.HandleFunc("/pages/search/", pages.SearchHandler)

	http.HandleFunc("/comments/create/", comments.MakeHandler(comments.CreateHandler))

	http.HandleFunc("/categories/create/", categories.CreateHandler)

	http.HandleFunc("/ticket/new", tickets.TicketNewHandler)
	http.HandleFunc("/ticket/view/", tickets.MakeHandler(tickets.TicketViewHandler))
	http.HandleFunc("/ticket/assign/", tickets.MakeHandler(tickets.TicketAssignHandler))
	http.HandleFunc("/ticket/complete/", tickets.MakeHandler(tickets.TicketCompleteHandler))
	http.HandleFunc("/ticket/queue", tickets.TicketQueueHandler)

	http.HandleFunc("/users/login/", users.LoginHandler)
	http.HandleFunc("/users/logout/", users.LogoutHandler)
	http.HandleFunc("/users/create/", users.CreateUserHandler)
	http.HandleFunc("/dashboard", pages.DashboardHandler)
	http.HandleFunc("/", RootHandler)

	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
}
