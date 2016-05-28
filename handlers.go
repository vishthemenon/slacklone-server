package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
)

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
  Id string `gorethink:"id,omitempty"`
  Name string `gorethink:"name"`
}

const (
  ChannelStop = iota
  UserStop
  MessageStop
)

func addChannel(client *Client, data interface{}) {
	var channel Channel
	if err := mapstructure.Decode(data, &channel); err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		if err := r.Table("channel").Insert(channel).Exec(client.session); err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := r.Table("channel").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
	}()
}

func editUser(client *Client, data interface{}) {
	var user User
	if err := mapstructure.Decode(data, &user); err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		if err := r.Table("user").Insert(user).Exec(client.session); err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func addUser(client *Client, data interface{}) {
	var user User
	if err := mapstructure.Decode(data, &user); err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		if err := r.Table("user").Insert(user).Exec(client.session); err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(UserStop)
		cursor, err := r.Table("user").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "user", client.send, stop)
	}()
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
