package main

import (
	"log"
	"net/http"
	"os"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
}

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)

type Message struct {
	Sender  *Client
	Content string
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}
	client := &Client{conn: conn}
	clients[client] = true

	defer func() {
		delete(clients, client)
		client.conn.Close()
	}()

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			return
		}
		broadcast <- message
	}
}

func handleMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			err := client.conn.WriteJSON(message)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}


func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {
	go handleMessages()
	http.HandleFunc("/", handleConnection)

	err := http.ListenAndServe(getPort(), nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
