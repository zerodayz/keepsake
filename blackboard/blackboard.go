package blackboard

import (
	"github.com/gorilla/websocket"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/zerodayz/keepsake/database"
	"github.com/zerodayz/keepsake/pages"
	"html/template"
	"math/rand"
	"net/http"
	"time"
)

var (
	username = ""
	templatePath = "tmpl/blackboard/"
)

type Client struct {
	id       string
	hub      *Hub
	color    string
	socket   *websocket.Conn
	outbound chan []byte
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateColor() string {
	c := colorful.Hsv(rand.Float64()*360.0, 0.8, 0.8)
	return c.Hex()
}

func newClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		id:       username,
		color:    generateColor(),
		hub:      hub,
		socket:   socket,
		outbound: make(chan []byte),
	}
}

func (client *Client) read() {
	defer func() {
		client.hub.unregister <- client
	}()
	for {
		_, data, err := client.socket.ReadMessage()
		if err != nil {
			break
		}
		client.hub.onMessage(data, client)
	}
}

func (client *Client) write() {
	for {
		select {
		case data, ok := <-client.outbound:
			if !ok {
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (client Client) run() {
	go client.read()
	go client.write()
}

func (client Client) close() {
	client.socket.Close()
	close(client.outbound)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	s := database.WikiPage{}
	t := template.Must(template.ParseFiles(templatePath + "create.html"))
	username = pages.ReadCookie(w, r)
	s.UserLoggedIn = username

	if username == "Unauthorized" {
		http.Redirect(w, r, "/users/login/", http.StatusFound)
		return
	}

	err := t.ExecuteTemplate(w, "create.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
