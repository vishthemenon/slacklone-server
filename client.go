package main

import (
  "log"
  "github.com/gorilla/websocket"
  r "github.com/dancannon/gorethink"
)
type FindHandler func (string) (Handler, bool)

type Message struct {
  Name string `json:"name"`
  Data interface{} `json:"data"`
}

type Client struct {
  send chan Message
  socket *websocket.Conn
  findHandler FindHandler
  session *r.Session
  stopChannels map[int]chan bool
  id string
  userName string
}

func (client *Client) Write() {
  for msg := range client.send {
    // fmt.Printf("%#v\n",msg)
    if err := client.socket.WriteJSON(msg); err !=nil {
      log.Println(err.Error())
      break
    }
  }
  client.socket.Close()
}
func (client *Client) Read() {
  var message Message
  for {
    if err := client.socket.ReadJSON(&message); err != nil {
      log.Println(err.Error())
      break
    }
    // what func to call
    if handler, found := client.findHandler(message.Name); found {
      handler(client, message.Data)
    }
  }
  client.socket.Close()
}

func (c *Client) NewStopChannel(stopKey int) chan bool {
  stop := make(chan bool)
  c.stopChannels[stopKey] = stop
  return stop
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
    var user User
    user.Name = "anonymous"
    res, err := r.Table("user").Insert(user).RunWrite(session)
    if err != nil {
      log.Println(err.Error())
    }
    var id string
    if len(res.GeneratedKeys) > 0 {
      id = res.GeneratedKeys[0]
    }
    return &Client {
    send: make(chan Message),
    socket: socket,
    findHandler: findHandler,
    session: session,
    stopChannels: make(map[int]chan bool),
    id: id,
    userName: user.Name,
  }
}
