package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	rdb "github.com/rethinkdb/rethinkdb-go"
)

func main() {
	session, err := rdb.Connect(rdb.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", os.Getenv("DB_URL"), "28015"),
		Database: os.Getenv("DB_NAME"),
	})
	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)

	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)

	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	router.Handle("message add", addChannelMessage)
	router.Handle("message subscribe", subscribeChannelMessage)
	router.Handle("message unsubscribe", unsubscribeChannelMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
