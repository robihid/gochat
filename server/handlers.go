package main

import (
	"time"

	"github.com/mitchellh/mapstructure"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
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

type Channel struct {
	ID   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
	ID   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

type ChannelMessage struct {
	ID        string    `gorethink:"id,omitempty"`
	ChannelID string    `gorethink:"channelID"`
	Body      string    `gorethink:"body"`
	Author    string    `gorethink:"author"`
	CreatedAt time.Time `gorethink:"createdAt"`
}

func editUser(c *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}
	c.userName = user.Name
	go func() {
		_, err := r.Table("user").
			Get(c.id).
			Update(user).
			RunWrite(c.session)
		if err != nil {
			c.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeUser(c *Client, data interface{}) {
	go func() {
		stop := c.NewStopChannel(UserStop)
		cursor, err := r.Table("user").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(c.session)

		if err != nil {
			c.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "user", c.send, stop)
	}()
}

func unsubscribeUser(c *Client, data interface{}) {
	c.StopForKey(UserStop)
}

func addChannelMessage(c *Client, data interface{}) {
	var channelMessage ChannelMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		c.send <- Message{"error", err.Error()}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.Author = c.userName
		err := r.Table("message").
			Insert(channelMessage).
			Exec(c.session)
		if err != nil {
			c.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannelMessage(c *Client, data interface{}) {
	go func() {
		eventData := data.(map[string]interface{})
		val, ok := eventData["channelID"]
		if !ok {
			return
		}
		channelID, ok := val.(string)
		if !ok {
			return
		}
		stop := c.NewStopChannel(MessageStop)
		cursor, err := r.Table("message").
			OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
			Filter(r.Row.Field("channelID").Eq(channelID)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(c.session)

		if err != nil {
			c.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "message", c.send, stop)
	}()
}

func unsubscribeChannelMessage(c *Client, data interface{}) {
	c.StopForKey(MessageStop)
}

func addChannel(c *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		c.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		err = r.Table("channel").
			Insert(channel).
			Exec(c.session)
		if err != nil {
			c.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannel(c *Client, data interface{}) {
	go func() {
		stop := c.NewStopChannel(ChannelStop)
		cursor, err := r.Table("channel").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(c.session)
		if err != nil {
			c.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", c.send, stop)
	}()
}

func unsubscribeChannel(c *Client, data interface{}) {
	c.StopForKey(ChannelStop)
}

func changeFeedHelper(cursor *r.Cursor, changeEventName string,
	send chan<- Message, stop <-chan bool) {
	change := make(chan r.ChangeResponse)
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
