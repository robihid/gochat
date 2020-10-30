package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	rdb "github.com/rethinkdb/rethinkdb-go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler func(*Client, interface{})

type Router struct {
	routes  map[string]Handler
	session *rdb.Session
}

func NewRouter(session *rdb.Session) *Router {
	return &Router{
		routes:  make(map[string]Handler),
		session: session,
	}
}

func (router *Router) Handle(msgName string, handler Handler) {
	router.routes[msgName] = handler
}

func (router *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := router.routes[msgName]
	return handler, found
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, router.FindHandler, router.session)
	defer client.Close()
	go client.Write()
	client.Read()
}
