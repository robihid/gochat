package main

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	rdb "github.com/rethinkdb/rethinkdb-go"
)

const (
	ChannelStop = iota
	UserStop
	MessageStop
)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type User struct {
	ID   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

type Channel struct {
	ID   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type ChannelMessage struct {
	ID        string    `gorethink:"id,omitempty"`
	ChannelID string    `gorethink:"channelId"`
	Body      string    `gorethink:"body"`
	Author    string    `gorethink:"author"`
	CreatedAt time.Time `gorethink:"createdAt"`
}

func addChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		err = rdb.Table("channel").
			Insert(channel).
			Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := rdb.Table("channel").Changes(rdb.ChangesOpts{IncludeInitial: true}).Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
	}()
}

func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

func editUser(client *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	client.username = user.Name
	go func() {
		_, err := rdb.Table("user").
			Get(client.id).
			Update(user).
			RunWrite(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(UserStop)
		cursor, err := rdb.Table("user").Changes(rdb.ChangesOpts{IncludeInitial: true}).Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "user", client.send, stop)
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addChannelMessage(client *Client, data interface{}) {
	fmt.Println("addChannelMessage called")
	var channelMessage ChannelMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.Author = client.username
		err := rdb.Table("message").
			Insert(channelMessage).
			Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannelMessage(client *Client, data interface{}) {
	go func() {
		eventData := data.(map[string]interface{})
		val, ok := eventData["channelId"]
		if !ok {
			return
		}
		channelId, ok := val.(string)
		if !ok {
			return
		}
		stop := client.NewStopChannel(MessageStop)
		cursor, err := rdb.Table("message").
			OrderBy(rdb.OrderByOpts{Index: rdb.Desc("createdAt")}).
			Filter(rdb.Row.Field("channelId").Eq(channelId)).
			Changes(rdb.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			fmt.Println("subscribeChannelMessage:", err.Error())
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func unsubscribeChannelMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func changeFeedHelper(cursor *rdb.Cursor, changeEventName string, send chan<- Message, stop <-chan bool) {
	change := make(chan rdb.ChangeResponse)
	cursor.Listen(change)
	for {
		eventName := ""
		var data interface{}
		select {
		case <-stop:
			cursor.Close()
			return
		case val := <-change:
			if val.NewValue != nil && val.OldValue == nil {
				eventName = changeEventName + " add"
				data = val.NewValue
			} else if val.NewValue == nil && val.OldValue != nil {
				eventName = changeEventName + " remove"
				data = val.OldValue
			} else if val.NewValue != nil && val.OldValue != nil {
				eventName = changeEventName + " edit"
				data = val.NewValue
			}
			send <- Message{eventName, data}
		}
	}
}
