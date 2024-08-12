package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	connections map[*websocket.Conn]bool
	sync.RWMutex
}

func NewServer() *Server {
	return &Server{connections: make(map[*websocket.Conn]bool), RWMutex: sync.RWMutex{}}
}

func (s *Server) handleWs(conn *websocket.Conn) {
	s.Lock()
	fmt.Println("new connection has been established by addr:", conn.RemoteAddr().String())

	s.connections[conn] = true
	s.Unlock()

	s.readLoop(conn)
}

func (s *Server) handleHeartBeatWs(c *websocket.Conn) {

	for {
		_, err := c.Write([]byte(fmt.Sprintf("your connection is alive at %s", time.Now().UTC().String())))
		if err != nil {
			fmt.Println("connection has been dropped for:", c.RemoteAddr().String())
			c.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *Server) readLoop(conn *websocket.Conn) {
	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err.Error())
			continue
		}

		msg := buff[:n]
		fmt.Println(string(msg))

		go s.broadcast(msg)

	}
}

func (s *Server) broadcast(b []byte) {
	for v, _ := range s.connections {
		_, err := v.Write(b)
		if err != nil {
			fmt.Println("write error:", err.Error())
		}

	}
}

func main() {

	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWs))
	http.Handle("/ws-heart-beat", websocket.Handler(server.handleHeartBeatWs))

	http.ListenAndServe(":3002", nil)
}
