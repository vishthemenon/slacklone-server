package main

import (
  "fmt"
  "net/http"
  "github.com/gorilla/websocket"
  r "github.com/dancannon/gorethink"
)

type Router struct {
  rules map[string]Handler
  session *r.Session
}

func NewRouter(session *r.Session) *Router{
  return &Router{
    rules: make(map[string] Handler),
    session: session,
  }
}

type Handler func(*Client, interface{})

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
  CheckOrigin:     func(r *http.Request) bool { return true },
}

func (r *Router) FindHandler(msgName string) (Handler, bool){
  handler, found := r.rules[msgName]
  return handler, found
}

func (r *Router) Handle(msgName string, handler Handler) {
  r.rules[msgName] = handler
}

func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  socket, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprint(w, err.Error())
    return
  }
  client := NewClient(socket, e.FindHandler, e.session)
  // var msg Message
  // msg.Name = "user id"
  // msg.Data = make(map[string]string)
  // msg.Data["id"] = client.id
  client.send <- Message{"user id", client.id}
  fmt.Printf("%#v\n",Message{"user id", client.id})
  go client.Write()
  client.Read()
}
