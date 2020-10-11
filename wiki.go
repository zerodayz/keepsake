// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"github.com/zerodayz/keepsake/categories"
	"github.com/zerodayz/keepsake/comments"
	"github.com/zerodayz/keepsake/database"
	"github.com/zerodayz/keepsake/pages"
	"github.com/zerodayz/keepsake/tickets"
	"github.com/zerodayz/keepsake/users"
	"log"
	"net/http"
	"os"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

var (
	noSsl bool = false
	key string = "./certs/server.key"
	cert string = "./certs/server.crt"
)

func init() {
	flag.BoolVar(&noSsl, "no-ssl", LookupEnvOrBool("KEEPSAKE_DISABLE_SSL", noSsl), "Disable SSL")
	flag.StringVar(&key, "key", LookupEnvOrString("KEEPSAKE_SSL_KEY", key), "SSL Key")
	flag.StringVar(&cert, "cert", LookupEnvOrString("KEEPSAKE_SSL_CERT", cert), "SSL Cert")
	flag.Parse()
}

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if val == "1" {
			return true
		}
	}
	return defaultVal
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	database.InitializeDatabase()

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("lib"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	http.HandleFunc("/revisions/view/", pages.MakeHandler(pages.RevisionsViewHandler))
	http.HandleFunc("/revisions/rollback/", pages.MakeHandler(pages.RevisionRollbackHandler))
	http.HandleFunc("/preview/view/", pages.MakeHandler(pages.PreviewHandler))
	http.HandleFunc("/preview/create/", pages.MakeHandler(pages.PreviewCreateHandler))
	http.HandleFunc("/pages/view/", pages.MakeHandler(pages.ViewHandler))
	http.HandleFunc("/pages/like/", pages.MakeHandler(pages.LikeHandler))
	http.HandleFunc("/pages/repair/", pages.MakeHandler(pages.RepairHandler))
	http.HandleFunc("/pages/download/", pages.DownloadHandler)
	http.HandleFunc("/pages/view/raw/", pages.ViewRawHandler)
	http.HandleFunc("/pages/stars", pages.StarHandler)
	http.HandleFunc("/pages/repairs/", pages.ListRepairsHandler)
	http.HandleFunc("/pages/list", pages.ListHandler)
	http.HandleFunc("/pages/revisions/", pages.MakeHandler(pages.RevisionsHandler))
	http.HandleFunc("/pages/edit/", pages.MakeHandler(pages.EditHandler))
	http.HandleFunc("/pages/delete/", pages.MakeHandler(pages.DeleteHandler))
	http.HandleFunc("/pages/create/", pages.CreateHandler)
	http.HandleFunc("/pages/upload/", pages.UploadFile)
	http.HandleFunc("/pages/save/", pages.MakeHandler(pages.SaveHandler))
	http.HandleFunc("/pages/trash/", pages.RecycleBinHandler)
	http.HandleFunc("/pages/restore/", pages.MakeHandler(pages.RestoreHandler))
	http.HandleFunc("/pages/search/", pages.SearchHandler)
	http.HandleFunc("/pages/search/raw/", pages.SearchRawHandler)

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
	if noSsl == true {
		log.Println("Starting Keepsake server at :80")
		log.Fatal(http.ListenAndServe(":80", nil))
	} else if noSsl == false {
		log.Println("Starting Keepsake server at :80")
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))
		log.Println("Starting Keepsake server at :443")
		log.Println("Serving SSL Key:", key, "and SSL Cert:", cert)
		log.Fatal(http.ListenAndServeTLS(":443", cert, key, nil))
	}
}
