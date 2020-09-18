package tickets

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zerodayz/keepsake/pages"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	templatePath = "tmpl/tickets/"
)

// Struct

type Page struct {
	DisplayBody  template.HTML
	UserLoggedIn string
}

type New struct {
	NowServing    int
	InQueue       int
	EstimatedTime string
}

type Queue struct {
	InternalId    int
	Name          string
	CompletedDate string
	CreatedDate   string
	Question      string
	Assigned      string
	Status        string
}

// DB

func AssignTicket(w http.ResponseWriter, r *http.Request, InternalId int, username string) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	PageUpdate, err := db.Prepare(`
	UPDATE queue SET status = ?, assigned = ?
	WHERE internal_id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = PageUpdate.Exec("Assigned", username, InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	http.Redirect(w, r, "/ticket/queue", http.StatusFound)
}

func CompleteTicket(w http.ResponseWriter, r *http.Request, InternalId int) {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	PageUpdate, err := db.Prepare(`
	UPDATE queue SET status = ?
	WHERE internal_id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = PageUpdate.Exec("Completed", InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	http.Redirect(w, r, "/ticket/queue", http.StatusFound)
}

func CreateTicket(w http.ResponseWriter, r *http.Request, q Queue) int {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	TicketInsert, err := db.Prepare(`
	INSERT INTO queue (name, question, status, date_created) VALUES ( ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal(err)
	}
	q.Status = "New"
	var res sql.Result
	res, err = TicketInsert.Exec(q.Name, q.Question, q.Status, q.CreatedDate)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	ticketId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return int(ticketId)

}

func ShowTicket(w http.ResponseWriter, r *http.Request, InternalId int) *Queue {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		name          string
		question      string
		dateCreated   string
		dateCompleted string
		assigned      string
		status        string
	)
	err = db.QueryRow(`
		SELECT name, question, date_created, date_completed, COALESCE(assigned, '') as assigned, status FROM queue WHERE internal_id = ?
		`, InternalId).Scan(&name, &question, &dateCreated, &dateCompleted, &assigned, &status)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return &Queue{InternalId: 0}
	}
	return &Queue{InternalId: InternalId, Name: name, Question: question, CreatedDate: dateCreated, CompletedDate: dateCompleted, Assigned: assigned, Status: status}
}

func FetchQueue(w http.ResponseWriter, r *http.Request) []Queue {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		queue         []Queue
		id            int
		name          string
		question      string
		dateCreated   string
		dateCompleted string
		assigned      string
		status        string
	)
	rows, err := db.Query("SELECT internal_id, name, question, date_created, date_completed, COALESCE(assigned, '') as assigned, status FROM queue ORDER BY internal_id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name, &question, &dateCreated, &dateCompleted, &assigned, &status)
		if err != nil {
			log.Fatal(err)
		}
		queue = append(queue, Queue{InternalId: id, Name: name, Question: question, CreatedDate: dateCreated, CompletedDate: dateCompleted, Assigned: assigned, Status: status})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return queue
}

func FetchNewQueue(w http.ResponseWriter, r *http.Request) []Queue {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		queue         []Queue
		id            int
		name          string
		question      string
		dateCreated   string
		dateCompleted string
		assigned      string
		status        string
	)
	rows, err := db.Query("SELECT internal_id, name, question, date_created, date_completed, COALESCE(assigned, '') as assigned, status FROM queue WHERE status = ? ORDER BY internal_id", "New")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name, &question, &dateCreated, &dateCompleted, &assigned, &status)
		if err != nil {
			log.Fatal(err)
		}
		queue = append(queue, Queue{InternalId: id, Name: name, Question: question, CreatedDate: dateCreated, CompletedDate: dateCompleted, Assigned: assigned, Status: status})
	}
	err = rows.Err()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	return queue
}

// Func

func secondsToDuration(inSeconds int) string {
	hours := inSeconds / 3600
	minutes := (inSeconds - (hours * 3600)) / 60
	seconds := inSeconds - (hours * 3600) - (minutes * 60)
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func LoadPage(w http.ResponseWriter, r *http.Request, InternalId int) (*Queue, error) {
	q := ShowTicket(w, r, InternalId)
	return q, nil
}

// Handlers

func TicketNewHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(templatePath + "new.html"))
	q := Queue{}
	p := &Page{}
	newQ := New{}

	username := pages.ReadCookie(w, r)

	if r.Method == "POST" {
		r.ParseForm()
		q.Name = r.PostFormValue("ircname")
		q.Question = r.PostFormValue("question")
		date := time.Now().UTC()
		q.CreatedDate = date.Format("20060102150405")

		q.InternalId = CreateTicket(w, r, q)
		http.Redirect(w, r, "/ticket/view/"+strconv.Itoa(q.InternalId), http.StatusFound)

	}

	queue := FetchQueue(w, r)
	newQueue := FetchNewQueue(w, r)

	newQ.NowServing = 0
	var (
		totalTimeInSecondsToResolution int
		numOfCompletedTickets          int
	)

	for _, i := range queue {
		if i.Status == "Completed" {
			dateCreated, err := time.Parse("2006-01-02 15:04:05", i.CreatedDate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			dateCompleted, err := time.Parse("2006-01-02 15:04:05", i.CompletedDate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			timeInSecondsToResolution := dateCompleted.Sub(dateCreated).Seconds()
			numOfCompletedTickets = +1
			totalTimeInSecondsToResolution = +int(timeInSecondsToResolution)
		}
		if i.Status == "Assigned" {
			newQ.NowServing = i.InternalId
		}
	}
	if numOfCompletedTickets != 0 {
		newQ.EstimatedTime = secondsToDuration(totalTimeInSecondsToResolution / numOfCompletedTickets)
	} else {
		newQ.EstimatedTime = "00:00:00"
	}
	newQ.InQueue = len(newQueue)
	p.UserLoggedIn = username

	err := t.ExecuteTemplate(w, "new.html", struct{ Page, New interface{} }{p, newQ})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TicketViewHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	t := template.Must(template.ParseFiles(templatePath + "view.html"))
	p := &Page{}

	username := pages.ReadCookie(w, r)

	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	q, err := LoadPage(w, r, id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	p.UserLoggedIn = username

	err = t.ExecuteTemplate(w, "view.html", struct{ Page, Queue interface{} }{p, q})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TicketCompleteHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	username := pages.ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	CompleteTicket(w, r, id)
}

func TicketAssignHandler(w http.ResponseWriter, r *http.Request, InternalId string) {
	username := pages.ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(InternalId)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	AssignTicket(w, r, id, username)
}

func TicketQueueHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(templatePath + "queue.html"))
	p := &Page{}

	username := pages.ReadCookie(w, r)
	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}

	buf := bytes.NewBuffer(nil)
	queue := FetchQueue(w, r)
	newQueue := FetchNewQueue(w, r)

	buf.Write([]byte(`<div>There are ` + strconv.Itoa(len(newQueue)) + ` tickets in the queue.</div>`))
	for _, f := range queue {
		if f.Status == "New" || f.Status == "Assigned" {
			buf.Write([]byte(`
				<div class="found">Created on ` + f.CreatedDate + ` by ` + f.Name +
				` | Change status to <a href="/ticket/assign/` + strconv.Itoa(f.InternalId) + `">Assigned</a> or <a href="/ticket/complete/` + strconv.Itoa(f.InternalId) + `">Completed</a>
				<label for="search-content" class="search-collapsible"> Ticket ` + strconv.Itoa(f.InternalId) + ` Status: ` + f.Status + `, Assigned: ` + f.Assigned + ` </label><div id="search-content" class="search-content">
				<pre><code>` + f.Question + `</code></pre></div></div>`))
		}
	}

	p.DisplayBody = template.HTML(buf.String())
	p.UserLoggedIn = username

	err := t.ExecuteTemplate(w, "queue.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(ticket)/(view|new|queue|assign|complete)/([0-9]+)$")

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
