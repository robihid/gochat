package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

type Handler func(*Client, interface{})

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Router struct {
	rules   map[string]Handler
	session *r.Session
}

func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

func (ro *Router) Handle(msgName string, handler Handler) {
	ro.rules[msgName] = handler
}

func (ro *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := ro.rules[msgName]
	return handler, found
}

func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, ro.FindHandler, ro.session)
	defer client.Close()
	go client.Write()
	client.Read()

}
