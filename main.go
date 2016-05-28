package main

import (
	"fmt"
	"log"
	"net/http"
	r "github.com/dancannon/gorethink"
)

var port = ":3000"

func main() {
	fmt.Println("Serving a hot pile of crap on localhost" + port)
	session, err := r.Connect(r.ConnectOpts{
		Address: "localhost:28025",
		Database: "slacklone",
	})
	if err != nil {
		log.Panic(err.Error())
	}
	router := NewRouter(session)
	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("user add", addUser)
	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	http.Handle("/", router)
	http.ListenAndServe(port, nil)
}
