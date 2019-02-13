package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang/net/websocket"
)

/*
func main() {
	l, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go match(conn)
	}
}
*/

const listenAddr = "localhost:8080"

func main() {
	http.HandleFunc("/", handler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, listenAddr)
}

type socket struct {
	io.ReadWriteCloser
	done chan bool
}

func (s *socket) Close() error {
	log.Printf("connection %p closed", s)
	s.done <- true
	return nil
}

func socketHandler(conn *websocket.Conn) {
	s := socket{conn, make(chan bool)}
	go match(&s)
	<-s.done
}

var partners = make(chan io.ReadWriteCloser)

func match(conn io.ReadWriteCloser) {
	fmt.Fprintln(conn, "Looking for a partner ...")
	select {
	case partners <- conn:
		// the other goroutine won and we can finish
	case p := <-partners:
		chat(conn, p)
	}
}

func chat(a, b io.ReadWriteCloser) {
	fmt.Fprintln(a, "We found a partner")
	fmt.Fprintln(b, "We found a partner")
	errc := make(chan error, 1)
	go copy(a, b, errc)
	go copy(b, a, errc)
	if err := <-errc; err != nil {
		log.Printf("Error chatting: %v", err)
	}
	a.Close()
	b.Close()
}

func copy(a, b io.ReadWriteCloser, errc chan<- error) {
	_, err := io.Copy(a, b)
	errc <- err
}
