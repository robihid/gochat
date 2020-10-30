package main

import (
	"log"

	"github.com/gorilla/websocket"
	rdb "github.com/rethinkdb/rethinkdb-go"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *rdb.Session
	stopChannels map[int]chan bool
	id           string
	username     string
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *rdb.Session) *Client {
	var user User
	user.Name = "anonymous"
	res, err := rdb.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		username:     user.Name,
	}
}

func (client *Client) StopForKey(key int) {
	if ch, found := client.stopChannels[key]; found {
		ch <- true
		delete(client.stopChannels, key)
	}
}

func (client *Client) NewStopChannel(key int) chan bool {
	client.StopForKey(key)
	stop := make(chan bool)
	client.stopChannels[key] = stop
	return stop
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}

func (client *Client) Write() {
	for message := range client.send {
		if err := client.socket.WriteJSON(message); err != nil {
			break
		}
	}
	client.socket.Close()
}

func (client *Client) Close() {
	for _, ch := range client.stopChannels {
		ch <- true
	}
	close(client.send)
	rdb.Table("user").Get(client.id).Delete().Exec(client.session)
}
